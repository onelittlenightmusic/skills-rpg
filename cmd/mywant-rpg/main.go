package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

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

func newRPGClient(base string) *rpgClient {
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

func printResult(v any, status int, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "server error (HTTP %d):\n", status)
		printJSON(v)
		os.Exit(1)
	}
	printJSON(v)
}

func printJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(v)
		return
	}
	fmt.Println(string(b))
}

func client() *rpgClient {
	return newRPGClient(serverURL)
}

var rootCmd = &cobra.Command{
	Use:     "mywant-rpg",
	Version: version,
	Short:   "MyWant RPG plugin - control the skills-rpg game server",
	Long: `mywant-rpg is a CLI plugin for MyWant that lets you control the skills-rpg game server.

The plugin connects to rpg-server (default: http://localhost:7100).
Set RPG_SERVER_URL environment variable or use --server-url to override.

Usage via mywant:
  mywant rpg start
  mywant rpg goal
  mywant rpg observe you
  mywant rpg saves
  mywant rpg save quicksave`,
}

// start: GET /api/v1/start
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Show role, rules, and current game state (call first each session)",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().get("/api/v1/start")
		printResult(v, status, err)
	},
}

// goal: GET /api/v1/next-goal
var goalCmd = &cobra.Command{
	Use:   "goal",
	Short: "Get the next suggested goal (text + hint + recommended skill)",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().get("/api/v1/next-goal")
		printResult(v, status, err)
	},
}

// observe [path]: GET /api/v1/observe
var observeCmd = &cobra.Command{
	Use:   "observe [path]",
	Short: "Read current game state (empty = full, or dot-path like 'you')",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "/api/v1/observe?actor=chap"
		if len(args) > 0 && args[0] != "" {
			path += "&target=" + urlEscape(args[0])
		}
		v, status, err := client().get(path)
		printResult(v, status, err)
	},
}

// control <action> [target]: POST /api/v1/control
var (
	controlActor string
	controlArgs  []string
)

var controlCmd = &cobra.Command{
	Use:   "control <action> [target]",
	Short: "Perform a game action (open, move, pickup, observe, ...)",
	Args:  cobra.RangeArgs(1, 2),
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
		if len(controlArgs) > 0 {
			extra := map[string]any{}
			for _, kv := range controlArgs {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					extra[parts[0]] = parts[1]
				}
			}
			payload["args"] = extra
		}
		v, status, err := client().send("POST", "/api/v1/control", payload)
		printResult(v, status, err)
	},
}

// saves: GET /api/v1/saves
var savesCmd = &cobra.Command{
	Use:   "saves",
	Short: "List all save slots",
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().get("/api/v1/saves")
		printResult(v, status, err)
	},
}

// save <slot> [--name label]: POST /api/v1/saves/{slot}
var saveLabel string

var saveCmd = &cobra.Command{
	Use:   "save <slot>",
	Short: "Save current game to a slot (e.g. quicksave, autosave, or custom name)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		payload := map[string]any{"name": saveLabel}
		v, status, err := client().send("POST", "/api/v1/saves/"+args[0], payload)
		printResult(v, status, err)
	},
}

// load <slot>: POST /api/v1/saves/{slot}/load
var loadCmd = &cobra.Command{
	Use:   "load <slot>",
	Short: "Load game from a save slot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().send("POST", "/api/v1/saves/"+args[0]+"/load", nil)
		printResult(v, status, err)
	},
}

// rm <slot>: DELETE /api/v1/saves/{slot}
var rmCmd = &cobra.Command{
	Use:     "rm <slot>",
	Aliases: []string{"delete"},
	Short:   "Delete a save slot",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().send("DELETE", "/api/v1/saves/"+args[0], nil)
		printResult(v, status, err)
	},
}

// debug subcommand group
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug commands (testing only)",
}

var debugJumpCmd = &cobra.Command{
	Use:   "jump <stage>",
	Short: "Teleport to the start of a stage (clears achievements and inventory)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v, status, err := client().send("POST", "/api/v1/debug/jump", map[string]any{"stage": args[0]})
		printResult(v, status, err)
	},
}

// server subcommand group
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the rpg-server process",
}

var serverBin string
var serverPort int
var serverDataDir string
var serverStagesDir string
var serverReset bool

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the rpg-server process in the background",
	Run: func(cmd *cobra.Command, args []string) {
		bin := serverBin
		if bin == "" {
			var err error
			bin, err = exec.LookPath("rpg-server")
			if err != nil {
				// Try skills-rpg bin directory
				home, _ := os.UserHomeDir()
				candidates := []string{
					filepath.Join(home, "work", "skills-rpg", "bin", "rpg-server"),
					"/usr/local/bin/rpg-server",
				}
				for _, c := range candidates {
					if _, err2 := os.Stat(c); err2 == nil {
						bin = c
						break
					}
				}
			}
			if bin == "" {
				fmt.Fprintln(os.Stderr, "rpg-server not found in PATH. Use --bin to specify the path.")
				os.Exit(1)
			}
		}

		cmdArgs := []string{}
		if serverPort != 7100 {
			cmdArgs = append(cmdArgs, "--port", strconv.Itoa(serverPort))
		}
		if serverDataDir != "" {
			cmdArgs = append(cmdArgs, "--data-dir", serverDataDir)
		}
		if serverStagesDir != "" {
			cmdArgs = append(cmdArgs, "--stages-dir", serverStagesDir)
		}
		if serverReset {
			cmdArgs = append(cmdArgs, "--reset")
		}

		pidFile := rpgPIDFile()
		logFile := rpgLogFile()

		if pid := readPID(pidFile); pid > 0 {
			if isRunning(pid) {
				fmt.Printf("rpg-server is already running (PID %d)\n", pid)
				return
			}
		}

		logF, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "cannot open log file:", err)
			os.Exit(1)
		}
		defer logF.Close()

		proc := exec.Command(bin, cmdArgs...)
		proc.Stdout = logF
		proc.Stderr = logF
		if err := proc.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "failed to start rpg-server:", err)
			os.Exit(1)
		}

		pid := proc.Process.Pid
		if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "warning: could not write PID file:", err)
		}

		fmt.Printf("rpg-server started (PID %d), log: %s\n", pid, logFile)
	},
}

var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the rpg-server process",
	Run: func(cmd *cobra.Command, args []string) {
		pidFile := rpgPIDFile()
		pid := readPID(pidFile)
		if pid <= 0 {
			fmt.Println("rpg-server is not running (no PID file)")
			return
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			fmt.Fprintln(os.Stderr, "cannot find process:", err)
			os.Exit(1)
		}
		if err := proc.Signal(os.Interrupt); err != nil {
			fmt.Fprintln(os.Stderr, "cannot signal process:", err)
			os.Exit(1)
		}
		_ = os.Remove(pidFile)
		fmt.Printf("rpg-server stopped (PID %d)\n", pid)
	},
}

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if rpg-server is running",
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

func rpgPIDFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mywant", "rpg-server.pid")
}

func rpgLogFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mywant", "rpg-server.log")
}

func readPID(path string) int {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	return pid
}

func isRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func urlEscape(s string) string {
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	defaultURL := "http://localhost:7100"
	if v := os.Getenv("RPG_SERVER_URL"); v != "" {
		defaultURL = v
	}

	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", defaultURL, "rpg-server base URL ($RPG_SERVER_URL)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output raw JSON")

	controlCmd.Flags().StringVar(&controlActor, "actor", "chap", "actor: 'chap' (AI agent) or 'you' (player)")
	controlCmd.Flags().StringArrayVar(&controlArgs, "arg", nil, "extra args as key=value pairs")

	saveCmd.Flags().StringVar(&saveLabel, "name", "", "human-readable label for the save slot")

	serverStartCmd.Flags().StringVar(&serverBin, "bin", "", "path to rpg-server binary (auto-detected if not set)")
	serverStartCmd.Flags().IntVar(&serverPort, "port", 7100, "port to listen on")
	serverStartCmd.Flags().StringVar(&serverDataDir, "data-dir", "", "data directory override")
	serverStartCmd.Flags().StringVar(&serverStagesDir, "stages-dir", "", "stages directory override")
	serverStartCmd.Flags().BoolVar(&serverReset, "reset", false, "reset game state on start")

	debugCmd.AddCommand(debugJumpCmd)
	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverStatusCmd)

	rootCmd.AddCommand(
		startCmd,
		goalCmd,
		observeCmd,
		controlCmd,
		savesCmd,
		saveCmd,
		loadCmd,
		rmCmd,
		debugCmd,
		serverCmd,
	)
}
