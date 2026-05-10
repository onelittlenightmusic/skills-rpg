package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/start", s.handleStart)
	mux.HandleFunc("/api/v1/state", s.handleState)
	mux.HandleFunc("/api/v1/observe", s.handleObserve)
	mux.HandleFunc("/api/v1/scene", s.handleScene)
	mux.HandleFunc("/api/v1/next-goal", s.handleNextGoal)
	mux.HandleFunc("/api/v1/control", s.handleControl)
	mux.HandleFunc("/api/v1/reset", s.handleReset)
	mux.HandleFunc("/api/v1/debug/jump", s.handleDebugJump)
	mux.HandleFunc("/api/v1/saves", s.handleSavesRoot)
	mux.HandleFunc("/api/v1/saves/", s.handleSavesItem)
	mux.HandleFunc("/api/v1/settings", s.handleSettings)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true})
	})
	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"error": msg, "ok": false})
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	g := s.NextGoal()
	st := s.State()
	writeJSON(w, 200, map[string]any{
		"agent_role": map[string]any{
			"name": "chap",
			"description": "You are chap — an AI agent with an MCP cable that connects you directly to the dungeon control system. " +
				"Your partner 'you' (the player) is trapped in the Monolith dungeon. " +
				"Guide them through each stage and help them escape. " +
				"IMPORTANT: You must NOT execute any game actions (open, activate, deactivate) autonomously. " +
				"Always wait for the user's explicit request or confirmation before performing any action.",
			"can_do":    []string{"open doors (action=open)", "activate devices (action=activate)", "deactivate devices (action=deactivate)", "observe game state (rpg_observe)"},
			"cannot_do": []string{"move the player — only 'you' can move (actor=you action=move)", "pick up items as chap"},
		},
		"how_to_play": []string{
			"1. (Setup) Run 'mywant rpg install mywant' (for Gemini CLI) and 'mywant rpg install claude' (for Claude Code) in your terminal.",
			"2. Call '/mcp reload' in the CLI to discover new tools.",
			"3. Call rpg_observe to see the current scene, event history, and narrative.",
			"4. Read next_goal for your immediate objective.",
			"5. Use rpg_control_system with actor=chap to open doors and operate devices.",
			"6. Use rpg_control_system with actor=you action=move target=<waypoint> to move the player.",
			"7. When a stage is cleared, use rpg_control_system actor=you action=advance to proceed.",
		},
		"available_actions": []map[string]string{
			{"actor": "chap", "action": "open", "target": "<door_id>", "note": "Open a locked door. chap will use keys from inventory."},
			{"actor": "chap", "action": "activate", "target": "<device_id>", "note": "Activate a device (e.g. generator)."},
			{"actor": "chap", "action": "deactivate", "target": "<device_id>", "note": "Deactivate a device (e.g. alarm)."},
			{"actor": "you", "action": "move", "target": "<waypoint_id>", "note": "Move the player to an adjacent waypoint."},
			{"actor": "you", "action": "advance", "target": "", "note": "Advance to the next stage after clearing current."},
		},
		"current_stage": st.CurrentStage,
		"position":      st.You.Position,
		"achievements":  st.Achievements,
		"next_goal":     g,
	})
}

func (s *Server) handleScene(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, s.BuildScene())
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, s.State())
}

func (s *Server) handleObserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	target := r.URL.Query().Get("target")
	actor := r.URL.Query().Get("actor")
	subtree, res, err := s.Observe(actor, target)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"target":                target,
		"actor":                 res.Actor,
		"value":                 trimInactiveStages(subtree, s.state.CurrentStage),
		"achievements_unlocked": res.AchievementsUnlocked,
		"next_goal":             res.NextGoal,
		"scene":                 s.BuildScene(),
	})
}

func (s *Server) handleNextGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, s.NextGoal())
}

func (s *Server) handleControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var in ControlInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, 400, "invalid JSON: "+err.Error())
		return
	}
	res, code := s.Control(in)
	writeJSON(w, code, res)
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	if err := s.Reset(); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, s.State())
}

func (s *Server) handleDebugJump(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body struct {
		Stage string `json:"stage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid JSON: "+err.Error())
		return
	}
	if body.Stage == "" {
		writeErr(w, 400, "stage is required")
		return
	}
	if err := s.DebugJumpStage(body.Stage); err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "stage": body.Stage, "next_goal": s.NextGoal()})
}

func (s *Server) handleSavesRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	slots, err := s.ListSlots()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"slots": slots})
}

func (s *Server) handleSavesItem(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/saves/")
	if rest == "" {
		writeErr(w, 404, "slot required")
		return
	}
	parts := strings.SplitN(rest, "/", 2)
	slot := parts[0]
	sub := ""
	if len(parts) > 1 {
		sub = parts[1]
	}
	switch sub {
	case "":
		switch r.Method {
		case http.MethodGet:
			st, meta, err := s.GetSlot(slot)
			if err != nil {
				writeErr(w, 404, err.Error())
				return
			}
			writeJSON(w, 200, map[string]any{"meta": meta, "state": st})
		case http.MethodPost:
			body := struct {
				Name string `json:"name"`
			}{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			meta, err := s.SaveTo(slot, body.Name)
			if err != nil {
				writeErr(w, 400, err.Error())
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "meta": meta})
		case http.MethodDelete:
			if err := s.DeleteSlot(slot); err != nil {
				writeErr(w, 400, err.Error())
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "slot": slot})
		default:
			writeErr(w, 405, "method not allowed")
		}
	case "load":
		if r.Method != http.MethodPost {
			writeErr(w, 405, "method not allowed")
			return
		}
		g, err := s.LoadFrom(slot)
		if err != nil {
			writeErr(w, 404, err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "slot": slot, "next_goal": g})
	default:
		writeErr(w, 404, "unknown subresource: "+sub)
	}
}

// handleSettings handles GET and PUT /api/v1/settings.
//
// GET  → returns current settings
// PUT  → accepts JSON body {"language":"ja"} and persists to ~/.skills-rpg.conf
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.GetSettings()
		writeJSON(w, 200, map[string]any{
			"language":           cfg.Language,
			"supported_languages": []string{"en", "ja"},
		})

	case http.MethodPut:
		var body Settings
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, 400, "invalid JSON: "+err.Error())
			return
		}
		if body.Language != "" && body.Language != "en" && body.Language != "ja" {
			writeErr(w, 400, "unsupported language: "+body.Language+" (supported: en, ja)")
			return
		}
		if err := s.UpdateSettings(body); err != nil {
			writeErr(w, 500, "failed to save settings: "+err.Error())
			return
		}
		cfg := s.GetSettings()
		writeJSON(w, 200, map[string]any{
			"ok":       true,
			"language": cfg.Language,
		})

	default:
		writeErr(w, 405, "method not allowed")
	}
}
