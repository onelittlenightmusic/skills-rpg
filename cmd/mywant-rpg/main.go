package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/onelittlenightmusic/skills-rpg/server"
	"github.com/spf13/cobra"
)

var (
	serverURL string
	jsonOut   bool
	version   = "dev"
)

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
	req, err := http.NewRequest(method, c.base+path, body)
	if err != nil {
		return nil, 0, err
	}
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

func printResult(v any, status int, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "server error (HTTP %d):\n", status)
		fmt.Println(formatJSON(v))
		os.Exit(1)
	}
	fmt.Println(formatJSON(v))
}

func client() *rpgClient {
	return newClient(serverURL)
}

var rootCmd = &cobra.Command{
	Use:     "mywant-rpg",
	Version: version,
	Short:   "MyWant RPG plugin - control the skills-rpg game server",
}

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().get("/api/v1/start")
		printResult(v, status, err)
	},
}

var goalCmd = &cobra.Command{
	Use: "goal",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().get("/api/v1/next-goal")
		printResult(v, status, err)
	},
}

var observeCmd = &cobra.Command{
	Use: "observe [path]",
	Run: func(cmd *cobra.Command, args []string) {
		path := "/api/v1/observe?actor=chap"
		if len(args) > 0 && args[0] != "" {
			path += "&target=" + urlEscape(args[0])
		}
		v, status, err := client().get(path)
		printResult(v, status, err)
	},
}

var (
	controlActor string
	controlArgs  []string
)

var controlCmd = &cobra.Command{
	Use: "control <action> [target]",
	Run: func(cmd *cobra.Command, args []string) {
		actor := controlActor
		if actor == "" {
			actor = "chap"
		}
		payload := map[string]any{
			"actor":  actor,
			"action": args[0],
		}
		if len(args) > 1 {
			payload["target"] = args[1]
		}
		v, status, err := client().send("POST", "/api/v1/control", payload)
		printResult(v, status, err)
	},
}

var savesCmd = &cobra.Command{
	Use: "saves",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().get("/api/v1/saves")
		printResult(v, status, err)
	},
}

var saveLabel string
var saveCmd = &cobra.Command{
	Use: "save <slot>",
	Run: func(cmd *cobra.Command, args []string) {
		payload := map[string]any{"name": saveLabel}
		v, status, err := client().send("POST", "/api/v1/saves/"+args[0], payload)
		printResult(v, status, err)
	},
}

var loadCmd = &cobra.Command{
	Use: "load <slot>",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().send("POST", "/api/v1/saves/"+args[0]+"/load", nil)
		printResult(v, status, err)
	},
}

var rmCmd = &cobra.Command{
	Use: "rm <slot>",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().send("DELETE", "/api/v1/saves/"+args[0], nil)
		printResult(v, status, err)
	},
}

var debugCmd = &cobra.Command{Use: "debug"}

var debugJumpCmd = &cobra.Command{
	Use: "jump <stage>",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().send("POST", "/api/v1/debug/jump", map[string]any{"stage": args[0]})
		printResult(v, status, err)
	},
}

var serverCmd = &cobra.Command{Use: "server"}
var serverPort int
var serverDataDir string
var serverStagesDir string
var serverReset bool

var serverStartCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := server.Config{
			DataDir:   serverDataDir,
			StagesDir: serverStagesDir,
			Port:      serverPort,
		}
		s, err := server.NewServer(cfg)
		if err != nil {
			fmt.Printf("failed to init server: %v\n", err)
			return
		}

		// Write PID file for the current process
		os.WriteFile(rpgPIDFile(), []byte(strconv.Itoa(os.Getpid())), 0644)

		fmt.Printf("rpg-server started (PID %d) on port %d\n", os.Getpid(), serverPort)

		// Run server in blocking mode
		addr := fmt.Sprintf(":%d", serverPort)
		if err := http.ListenAndServe(addr, s.Handler()); err != nil {
			fmt.Printf("server error: %v\n", err)
		}
	},
}

var serverStopCmd = &cobra.Command{
	Use: "stop",
	Run: func(cmd *cobra.Command, args []string) {
		pidFile := rpgPIDFile()
		pid := readPID(pidFile)
		if pid <= 0 {
			fmt.Println("rpg-server is not running (no PID file)")
			return
		}
		proc, _ := os.FindProcess(pid)
		proc.Signal(os.Interrupt)
		_ = os.Remove(pidFile)
		fmt.Printf("rpg-server stopped (PID %d)\n", pid)
	},
}

var serverStatusCmd = &cobra.Command{
	Use: "status",
	Run: func(cmd *cobra.Command, args []string) {
		pidFile := rpgPIDFile()
		pid := readPID(pidFile)
		if pid <= 0 {
			fmt.Println("rpg-server: stopped")
			return
		}
		if isRunning(pid) {
			fmt.Printf("rpg-server: running (PID %d)\n", pid)
		} else {
			fmt.Println("rpg-server: stopped (stale PID file)")
			_ = os.Remove(pidFile)
		}
	},
}

var serveCmd = &cobra.Command{
	Use:    "_serve",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := server.Config{
			DataDir:   serverDataDir,
			StagesDir: serverStagesDir,
			Port:      serverPort,
		}
		s, _ := server.NewServer(cfg)
		addr := fmt.Sprintf(":%d", serverPort)
		http.ListenAndServe(addr, s.Handler())
	},
}

func init() {
	serveCmd.Flags().IntVar(&serverPort, "port", 7100, "server port")
	serveCmd.Flags().StringVar(&serverDataDir, "data-dir", os.ExpandEnv("$HOME/.mywant-rpg"), "data directory")
	serveCmd.Flags().StringVar(&serverStagesDir, "stages-dir", "", "stages directory")
}

var mcpCmd = &cobra.Command{Use: "mcp"}

var mcpServeCmd = &cobra.Command{
	Use: "serve",
	Run: func(cmd *cobra.Command, args []string) {
		runMCPServer(serverURL)
	},
}

func runMCPServer(apiURL string) {
	c := newClient(apiURL)
	serverInstance := mcp.NewServer(&mcp.Implementation{Name: "rpg-mcp", Version: "0.1.0"}, nil)

	mcp.AddTool(serverInstance, &mcp.Tool{
		Name: "rpg_start",
		Description: "** Call this first at the start of every session. **" +
			" Returns your role as chap (the AI agent), available actions, how to play, and current game state." +
			" Read this before doing anything else so you understand your mission and how to operate." +
			" IMPORTANT RULE: Never perform game actions (open doors, pick up items, move, etc.) autonomously." +
			" Always wait for an explicit instruction from the user before taking any action.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		v, code, err := c.get("/api/v1/start")
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	mcp.AddTool(serverInstance, &mcp.Tool{
		Name:        "rpg_next_goal",
		Description: "Return the player's next suggested goal (text + hint + recommended skill).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		v, code, err := c.get("/api/v1/next-goal")
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	mcp.AddTool(serverInstance, &mcp.Tool{
		Name:        "rpg_observe",
		Description: "Read a subtree of the game state by dot-path (e.g. 'you', 'stages.stage1.doors.door1'). Empty = full state. Records an observe event so look-style achievements fire.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Target string `json:"target,omitempty" jsonschema:"Dot-path target (omit or empty for full state)"`
	}) (*mcp.CallToolResult, any, error) {
		path := "/api/v1/observe?actor=chap"
		if input.Target != "" {
			path += "&target=" + urlEscape(input.Target)
		}
		v, code, err := c.get(path)
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	mcp.AddTool(serverInstance, &mcp.Tool{
		Name:        "rpg_control_system",
		Description: "Perform a game action. Use actor=\"you\" for player actions (move, observe) or actor=\"chap\" for AI agent actions (open doors, etc.). Defaults to \"chap\".",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Action string         `json:"action" jsonschema:"e.g. observe, move, pickup, open"`
		Target string         `json:"target,omitempty" jsonschema:"action target (waypoint, item, door, ...)"`
		Actor  string         `json:"actor,omitempty" jsonschema:"'you' or 'chap' (default: chap)"`
		Args   map[string]any `json:"args,omitempty" jsonschema:"optional extra arguments"`
	}) (*mcp.CallToolResult, any, error) {
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

	mcp.AddTool(serverInstance, &mcp.Tool{
		Name:        "rpg_save_list",
		Description: "List existing save slots with metadata.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		v, code, err := c.get("/api/v1/saves")
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	mcp.AddTool(serverInstance, &mcp.Tool{
		Name:        "rpg_save",
		Description: "Save current game state to a slot. Slot name is alphanumeric+_- (max 32 chars). Reserved names: autosave, quicksave.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Slot string `json:"slot" jsonschema:"Slot identifier"`
		Name string `json:"name,omitempty" jsonschema:"Optional human-readable label"`
	}) (*mcp.CallToolResult, any, error) {
		v, code, err := c.send("POST", "/api/v1/saves/"+input.Slot, map[string]any{"name": input.Name})
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	mcp.AddTool(serverInstance, &mcp.Tool{
		Name:        "rpg_load",
		Description: "Restore game state from a save slot. Returns the new next_goal.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Slot string `json:"slot" jsonschema:"Slot identifier"`
	}) (*mcp.CallToolResult, any, error) {
		v, code, err := c.send("POST", "/api/v1/saves/"+input.Slot+"/load", nil)
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	mcp.AddTool(serverInstance, &mcp.Tool{
		Name:        "rpg_save_delete",
		Description: "Delete a save slot.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Slot string `json:"slot" jsonschema:"Slot identifier"`
	}) (*mcp.CallToolResult, any, error) {
		v, code, err := c.send("DELETE", "/api/v1/saves/"+input.Slot, nil)
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	mcp.AddTool(serverInstance, &mcp.Tool{
		Name:        "rpg_debug_jump_stage",
		Description: "[DEBUG] Operator/test-only tool. Teleport to the start of any stage, clearing achievements and inventory. This is not an in-world chap action.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Stage string `json:"stage" jsonschema:"Stage ID to jump to (e.g. stage4)"`
	}) (*mcp.CallToolResult, any, error) {
		v, code, err := c.send("POST", "/api/v1/debug/jump", map[string]any{"stage": input.Stage})
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	serverInstance.Run(context.Background(), &mcp.StdioTransport{})
}

func wrap(v any, status int, err error) (*mcp.CallToolResult, any) {
	out := make(map[string]any)
	if m, ok := v.(map[string]any); ok {
		out = m
	} else if v != nil {
		out["data"] = v
	}
	if err != nil {
		out["error"] = err.Error()
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}, out
	}
	if status >= 400 {
		out["http_status"] = status
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: formatJSON(v)}}}, out
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: formatJSON(v)}}}, out
}

var installCmd = &cobra.Command{Use: "install"}
var installMyWantCmd = &cobra.Command{Use: "mywant", Run: func(cmd *cobra.Command, args []string) { runInstall("mywant") }}
var installClaudeCmd = &cobra.Command{Use: "claude", Run: func(cmd *cobra.Command, args []string) { runInstall("claude") }}
var installGeminiCmd = &cobra.Command{Use: "gemini", Run: func(cmd *cobra.Command, args []string) { runInstall("gemini") }}
var installCodexCmd = &cobra.Command{Use: "codex", Run: func(cmd *cobra.Command, args []string) { runInstall("codex") }}

var uninstallCmd = &cobra.Command{Use: "uninstall"}
var uninstallMyWantCmd = &cobra.Command{Use: "mywant", Run: func(cmd *cobra.Command, args []string) { runUninstall("mywant") }}
var uninstallClaudeCmd = &cobra.Command{Use: "claude", Run: func(cmd *cobra.Command, args []string) { runUninstall("claude") }}
var uninstallGeminiCmd = &cobra.Command{Use: "gemini", Run: func(cmd *cobra.Command, args []string) { runUninstall("gemini") }}
var uninstallCodexCmd = &cobra.Command{Use: "codex", Run: func(cmd *cobra.Command, args []string) { runUninstall("codex") }}

func runInstall(target string) {
	dst := getSkillPath(target)
	installSkillsTo(dst)
	fmt.Printf("Installed to %s\n", dst)
}

func runUninstall(target string) {
	dstBase := getSkillPath(target)
	entries, _ := fs.ReadDir(server.DefaultDataFS, "skills")
	for _, entry := range entries {
		if entry.IsDir() {
			os.RemoveAll(filepath.Join(dstBase, entry.Name()))
		}
	}
	fmt.Printf("Uninstalled from %s\n", dstBase)
}

func getSkillPath(target string) string {
	home, _ := os.UserHomeDir()
	switch target {
	case "mywant":
		return filepath.Join(home, ".mywant", "custom-types")
	case "claude":
		return filepath.Join(home, ".claude", "skills")
	case "gemini":
		return filepath.Join(home, ".gemini", "skills")
	case "codex":
		return filepath.Join(home, ".codex", "skills")
	default:
		return ""
	}
}

func installSkillsTo(dstBase string) error {
	entries, _ := fs.ReadDir(server.DefaultDataFS, "skills")
	os.MkdirAll(dstBase, 0755)
	for _, entry := range entries {
		if entry.IsDir() {
			dst := filepath.Join(dstBase, entry.Name())
			os.RemoveAll(dst)
			os.MkdirAll(dst, 0755)
			copyFS(server.DefaultDataFS, filepath.Join("skills", entry.Name()), dst)
		}
	}
	return nil
}

func copyFS(srcFS fs.FS, srcDir, dstDir string) error {
	return fs.WalkDir(srcFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			data, _ := fs.ReadFile(srcFS, path)
			rel, _ := filepath.Rel(srcDir, path)
			os.WriteFile(filepath.Join(dstDir, rel), data, 0644)
		}
		return nil
	})
}

func rpgPIDFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mywant", "rpg-server.pid")
}
func rpgLogFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mywant", "rpg-server.log")
}
func readPID(path string) int {
	b, _ := os.ReadFile(path)
	pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	return pid
}
func isRunning(pid int) bool {
	proc, _ := os.FindProcess(pid)
	return proc.Signal(syscall.Signal(0)) == nil
}
func urlEscape(s string) string {
	const hex = "0123456789ABCDEF"
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9', c == '-', c == '_', c == '.', c == '~':
			out = append(out, c)
		default:
			out = append(out, '%', hex[c>>4], hex[c&15])
		}
	}
	return string(out)
}

func main() {
	rootCmd.AddCommand(startCmd, goalCmd, observeCmd, controlCmd, savesCmd, saveCmd, loadCmd, rmCmd, debugCmd, serverCmd, serveCmd, installCmd, uninstallCmd, mcpCmd)
	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverStatusCmd)
	debugCmd.AddCommand(debugJumpCmd)
	installCmd.AddCommand(installMyWantCmd, installClaudeCmd, installGeminiCmd, installCodexCmd)
	uninstallCmd.AddCommand(uninstallMyWantCmd, uninstallClaudeCmd, uninstallGeminiCmd, uninstallCodexCmd)
	mcpCmd.AddCommand(mcpServeCmd)
	rootCmd.Execute()
}

func init() {
	defaultURL := "http://localhost:7100"
	if v := os.Getenv("RPG_SERVER_URL"); v != "" {
		defaultURL = v
	}
	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", defaultURL, "rpg-server base URL")

	// Use PersistentFlags on rootCmd so all subcommands (including hidden ones) can access them.
	rootCmd.PersistentFlags().IntVar(&serverPort, "port", 7100, "server port")
	rootCmd.PersistentFlags().StringVar(&serverDataDir, "data-dir", os.ExpandEnv("$HOME/.mywant-rpg"), "data directory")
	rootCmd.PersistentFlags().StringVar(&serverStagesDir, "stages-dir", "", "stages directory")
}
