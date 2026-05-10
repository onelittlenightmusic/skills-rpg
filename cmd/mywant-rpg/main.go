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
	"os/exec"
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
	Use:   "start",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().get("/api/v1/start")
		printResult(v, status, err)
	},
}

var goalCmd = &cobra.Command{
	Use:   "goal",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().get("/api/v1/next-goal")
		printResult(v, status, err)
	},
}

var observeCmd = &cobra.Command{
	Use:   "observe [path]",
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
	Use:   "control <action> [target]",
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
	Use:   "saves",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().get("/api/v1/saves")
		printResult(v, status, err)
	},
}

var saveLabel string
var saveCmd = &cobra.Command{
	Use:   "save <slot>",
	Run: func(cmd *cobra.Command, args []string) {
		payload := map[string]any{"name": saveLabel}
		v, status, err := client().send("POST", "/api/v1/saves/"+args[0], payload)
		printResult(v, status, err)
	},
}

var loadCmd = &cobra.Command{
	Use:   "load <slot>",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().send("POST", "/api/v1/saves/"+args[0]+"/load", nil)
		printResult(v, status, err)
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm <slot>",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().send("DELETE", "/api/v1/saves/"+args[0], nil)
		printResult(v, status, err)
	},
}

var debugCmd = &cobra.Command{Use: "debug"}

var debugJumpCmd = &cobra.Command{
	Use:   "jump <stage>",
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
	Use:   "start",
	Run: func(cmd *cobra.Command, args []string) {
		exe, _ := os.Executable()
		cmdArgs := []string{"_serve"}
		if serverPort != 7100 {
			cmdArgs = append(cmdArgs, "--port", strconv.Itoa(serverPort))
		}
		proc := exec.Command(exe, cmdArgs...)
		logF, _ := os.OpenFile(rpgLogFile(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		proc.Stdout = logF
		proc.Stderr = logF
		proc.Start()
		fmt.Printf("rpg-server started (PID %d)\n", proc.Process.Pid)
	},
}

var serverStopCmd = &cobra.Command{
	Use:   "stop",
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
	Use:   "status",
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

var mcpCmd = &cobra.Command{Use: "mcp"}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Run: func(cmd *cobra.Command, args []string) {
		runMCPServer(serverURL)
	},
}

func runMCPServer(apiURL string) {
	c := newClient(apiURL)
	serverInstance := mcp.NewServer(&mcp.Implementation{Name: "rpg-mcp", Version: "0.1.0"}, nil)

	mcp.AddTool(serverInstance, &mcp.Tool{Name: "rpg_start"}, func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		v, code, err := c.get("/api/v1/start")
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	mcp.AddTool(serverInstance, &mcp.Tool{Name: "rpg_observe"}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Target string `json:"target,omitempty"`
	}) (*mcp.CallToolResult, any, error) {
		path := "/api/v1/observe?actor=chap"
		if input.Target != "" {
			path += "&target=" + urlEscape(input.Target)
		}
		v, code, err := c.get(path)
		res, out := wrap(v, code, err)
		return res, out, nil
	})

	mcp.AddTool(serverInstance, &mcp.Tool{Name: "rpg_control_system"}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Action string         `json:"action"`
		Target string         `json:"target,omitempty"`
		Actor  string         `json:"actor,omitempty"`
		Args   map[string]any `json:"args,omitempty"`
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
	case "mywant": return filepath.Join(home, ".mywant", "custom-types")
	case "claude": return filepath.Join(home, ".claude", "skills")
	case "gemini": return filepath.Join(home, ".gemini", "skills")
	case "codex": return filepath.Join(home, ".codex", "skills")
	default: return ""
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

func rpgPIDFile() string { home, _ := os.UserHomeDir(); return filepath.Join(home, ".mywant", "rpg-server.pid") }
func rpgLogFile() string { home, _ := os.UserHomeDir(); return filepath.Join(home, ".mywant", "rpg-server.log") }
func readPID(path string) int { b, _ := os.ReadFile(path); pid, _ := strconv.Atoi(strings.TrimSpace(string(b))); return pid }
func isRunning(pid int) bool { proc, _ := os.FindProcess(pid); return proc.Signal(syscall.Signal(0)) == nil }
func urlEscape(s string) string {
	const hex = "0123456789ABCDEF"
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9', c == '-', c == '_', c == '.', c == '~': out = append(out, c)
		default: out = append(out, '%', hex[c>>4], hex[c&15])
		}
	}
	return string(out)
}

func main() {
	rootCmd.AddCommand(startCmd, goalCmd, observeCmd, controlCmd, savesCmd, saveCmd, loadCmd, rmCmd, debugCmd, serverCmd, serveCmd, installCmd, uninstallCmd, mcpCmd)
	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverStatusCmd)
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
}
