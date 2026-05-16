package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Output = map[string]any

type rpgClient struct {
	base string
	http *http.Client
}

func newClient(base string) *rpgClient {
	return &rpgClient{base: base, http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *rpgClient) get(path string) (any, int, error) {
	resp, err := c.http.Get(c.base + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var v any
	_ = json.Unmarshal(body, &v)
	return v, resp.StatusCode, nil
}

func (c *rpgClient) send(method, path string, payload any) (any, int, error) {
	var body io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		body = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.base+path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var v any
	_ = json.Unmarshal(respBody, &v)
	return v, resp.StatusCode, nil
}

func formatJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// wrap converts a server response into both a text-content CallToolResult and
// a structured Output map. Both are returned so MCP clients that surface only
// one of the two channels still see the data.
func wrap(v any, status int, err error) (*mcp.CallToolResult, Output) {
	out := Output{}
	if m, ok := v.(map[string]any); ok {
		out = m
	} else if v != nil {
		out["data"] = v
	}
	if err != nil {
		out["error"] = err.Error()
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		}, out
	}
	if status >= 400 {
		out["http_status"] = status
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: formatJSON(v)}},
		}, out
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: formatJSON(v)}},
	}, out
}

func main() {
	defaultURL := "http://localhost:7100"
	if v := os.Getenv("RPG_SERVER_URL"); v != "" {
		defaultURL = v
	}
	apiURL := flag.String("server-url", defaultURL, "rpg-server base URL")
	flag.Parse()

	log.SetOutput(os.Stderr)
	log.Printf("[rpg-mcp] starting, server=%s", *apiURL)

	c := newClient(*apiURL)
	server := mcp.NewServer(&mcp.Implementation{Name: "rpg-mcp", Version: "0.1.0"}, nil)

	// rpg-start
	mcp.AddTool(server, &mcp.Tool{
		Name: "rpg_start",
		Description: "** Call this first at the start of every session. **" +
			" Returns your role as chap (the AI agent), available actions, how to play, and current game state." +
			" Read this before doing anything else so you understand your mission and how to operate." +
			" IMPORTANT RULE: Never perform game actions (open doors, pick up items, move, etc.) autonomously." +
			" Always wait for an explicit instruction from the user before taking any action.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Output, error) {
		v, code, err := c.get("/api/v1/start")
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	// rpg-next-goal
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rpg_next_goal",
		Description: "Return the player's next suggested goal (text + hint + recommended skill).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Output, error) {
		v, code, err := c.get("/api/v1/next-goal")
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	// rpg-observe
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rpg_observe",
		Description: "Read a subtree of the game state by dot-path (e.g. 'you', 'stages.stage1.doors.door1'). Empty = full state. Records an observe event so look-style achievements fire.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Target string `json:"target,omitempty" jsonschema:"Dot-path target (omit or empty for full state)"`
	}) (*mcp.CallToolResult, Output, error) {
		path := "/api/v1/observe?actor=chap"
		if input.Target != "" {
			path += "&target=" + httpEscape(input.Target)
		}
		v, code, err := c.get(path)
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	// rpg-control-system
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rpg_control_system",
		Description: "Perform a game action. Use actor=\"you\" for player actions (move, observe) or actor=\"chap\" for AI agent actions (open doors, etc.). Defaults to \"chap\".",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Action string         `json:"action" jsonschema:"e.g. observe, move, pickup, open"`
		Target string         `json:"target,omitempty" jsonschema:"action target (waypoint, item, door, ...)"`
		Actor  string         `json:"actor,omitempty" jsonschema:"'you' or 'chap' (default: chap)"`
		Args   map[string]any `json:"args,omitempty" jsonschema:"optional extra arguments"`
	}) (*mcp.CallToolResult, Output, error) {
		actor := input.Actor
		if actor == "" {
			actor = "chap"
		}
		v, code, err := c.send("POST", "/api/v1/control", map[string]any{
			"actor": actor, "action": input.Action, "target": input.Target, "args": input.Args,
		})
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	// rpg-save-list
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rpg_save_list",
		Description: "List existing save slots with metadata.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, Output, error) {
		v, code, err := c.get("/api/v1/saves")
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	// rpg-save
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rpg_save",
		Description: "Save current game state to a slot. Slot name is alphanumeric+_- (max 32 chars). Reserved names: autosave, quicksave.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Slot string `json:"slot" jsonschema:"Slot identifier"`
		Name string `json:"name,omitempty" jsonschema:"Optional human-readable label"`
	}) (*mcp.CallToolResult, Output, error) {
		v, code, err := c.send("POST", "/api/v1/saves/"+input.Slot, map[string]any{"name": input.Name})
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	// rpg-load
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rpg_load",
		Description: "Restore game state from a save slot. Returns the new next_goal.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Slot string `json:"slot" jsonschema:"Slot identifier"`
	}) (*mcp.CallToolResult, Output, error) {
		v, code, err := c.send("POST", "/api/v1/saves/"+input.Slot+"/load", nil)
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	// rpg-save-delete
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rpg_save_delete",
		Description: "Delete a save slot.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Slot string `json:"slot" jsonschema:"Slot identifier"`
	}) (*mcp.CallToolResult, Output, error) {
		v, code, err := c.send("DELETE", "/api/v1/saves/"+input.Slot, nil)
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	// rpg-debug-jump-stage (debug only)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rpg_debug_jump_stage",
		Description: "[DEBUG] Teleport to the start of any stage, clearing achievements and inventory. For testing only.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Stage string `json:"stage" jsonschema:"Stage ID to jump to (e.g. stage4)"`
	}) (*mcp.CallToolResult, Output, error) {
		v, code, err := c.send("POST", "/api/v1/debug/jump", map[string]any{"stage": input.Stage})
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("rpg-mcp run: %v", err)
	}
}

// httpEscape: minimal URL escape for query values (no external dep needed).
func httpEscape(s string) string {
	const hex = "0123456789ABCDEF"
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9',
			c == '-', c == '_', c == '.', c == '~':
			out = append(out, c)
		default:
			out = append(out, '%', hex[c>>4], hex[c&15])
		}
	}
	return string(out)
}
