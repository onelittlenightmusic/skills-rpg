package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/state", s.handleState)
	mux.HandleFunc("/api/v1/observe", s.handleObserve)
	mux.HandleFunc("/api/v1/next-goal", s.handleNextGoal)
	mux.HandleFunc("/api/v1/control", s.handleControl)
	mux.HandleFunc("/api/v1/reset", s.handleReset)
	mux.HandleFunc("/api/v1/debug/jump", s.handleDebugJump)
	mux.HandleFunc("/api/v1/saves", s.handleSavesRoot)
	mux.HandleFunc("/api/v1/saves/", s.handleSavesItem)
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
		"value":                 subtree,
		"achievements_unlocked": res.AchievementsUnlocked,
		"next_goal":             res.NextGoal,
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
