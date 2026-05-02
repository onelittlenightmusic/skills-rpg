package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	DataDir   string // ~/.mywant-rpg
	StagesDir string // path to stages/ for bootstrap
	Port      int    // 7100
}

type Server struct {
	cfg   Config
	mu    sync.Mutex
	state *GameState
}

func NewServer(cfg Config) (*Server, error) {
	s := &Server{cfg: cfg}
	if err := s.loadOrBootstrap(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) currentPath() string { return filepath.Join(s.cfg.DataDir, "current.yaml") }

func (s *Server) loadOrBootstrap() error {
	if err := os.MkdirAll(s.cfg.DataDir, 0o755); err != nil {
		return err
	}
	var st GameState
	if err := readYAML(s.currentPath(), &st); err == nil {
		s.state = &st
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	gs, err := initialStateFromStages(s.cfg.StagesDir)
	if err != nil {
		return err
	}
	s.state = gs
	return s.persistLocked()
}

// Reset rebuilds state from stage YAMLs, discarding current.yaml.
func (s *Server) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	gs, err := initialStateFromStages(s.cfg.StagesDir)
	if err != nil {
		return err
	}
	s.state = gs
	return s.persistLocked()
}

func (s *Server) persistLocked() error {
	return writeYAMLAtomic(s.currentPath(), s.state)
}

// State returns a deep copy of the current state (via JSON round-trip) safe for
// JSON encoding without holding the lock.
func (s *Server) State() *GameState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneState(s.state)
}

// Observe returns the subtree at the given dot-path. Empty path returns whole state.
// Records an observe event (so look_around-style achievements fire).
func (s *Server) Observe(actor, target string) (any, *ControlResult, error) {
	if actor == "" {
		actor = ActorYou
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	tree, err := stateAsTree(s.state)
	if err != nil {
		return nil, nil, err
	}
	subtree, err := walkPath(tree, target)
	if err != nil {
		return nil, nil, err
	}

	ev := Event{Actor: actor, Action: ActionObserve, Target: target, Result: "ok"}
	res := finishResult(s.state, ev, ControlResult{
		OK: true, Actor: actor, Action: ActionObserve, Target: target,
	})
	if err := s.persistLocked(); err != nil {
		return nil, nil, err
	}
	return subtree, &res, nil
}

// Control applies a control input.
func (s *Server) Control(in ControlInput) (ControlResult, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, res := applyControl(s.state, in)
	if err := s.persistLocked(); err != nil {
		// Persistence failure is server-side; surface but keep result.
		res.Reason = fmt.Sprintf("%s; persist error: %v", res.Reason, err)
		return res, 500
	}
	if !res.OK {
		return res, 409
	}
	return res, 200
}

func (s *Server) NextGoal() Goal {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state.NextGoal
}

// DebugJumpStage teleports the player to the start of the given stage,
// clearing achievements and inventory. Only for testing/debug use.
func (s *Server) DebugJumpStage(stageID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stage, ok := s.state.Stages[stageID]
	if !ok {
		return fmt.Errorf("stage %q not found", stageID)
	}
	s.state.CurrentStage = stageID
	s.state.You.Position = stage.InitialPosition
	s.state.You.Inventory = nil
	s.state.Achievements = []string{}
	s.state.NextGoal = computeNextGoal(s.state)
	return s.persistLocked()
}

// --- saves ---

func (s *Server) SaveTo(slot, name string) (SaveMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeSlot(s.cfg.DataDir, slot, name, s.state)
}

func (s *Server) LoadFrom(slot string) (Goal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, _, err := readSlot(s.cfg.DataDir, slot)
	if err != nil {
		return Goal{}, err
	}
	s.state = st
	g := computeNextGoal(s.state)
	s.state.NextGoal = g
	if err := s.persistLocked(); err != nil {
		return Goal{}, err
	}
	return g, nil
}

func (s *Server) DeleteSlot(slot string) error {
	return deleteSlot(s.cfg.DataDir, slot)
}

func (s *Server) GetSlot(slot string) (*GameState, SaveMeta, error) {
	return readSlot(s.cfg.DataDir, slot)
}

func (s *Server) ListSlots() ([]SaveMeta, error) {
	return listSlots(s.cfg.DataDir)
}

// --- helpers ---

func cloneState(in *GameState) *GameState {
	b, _ := json.Marshal(in)
	var out GameState
	_ = json.Unmarshal(b, &out)
	return &out
}

func stateAsTree(in *GameState) (map[string]any, error) {
	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// walkPath descends a dot-separated path into a JSON-shaped tree.
// Supports `key`, `key.subkey`, and `key[N]` for array indices.
func walkPath(tree any, path string) (any, error) {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return tree, nil
	}
	cur := tree
	for _, segment := range strings.Split(path, ".") {
		key, idx, hasIdx := parseSegment(segment)
		switch v := cur.(type) {
		case map[string]any:
			next, ok := v[key]
			if !ok {
				return nil, fmt.Errorf("path %q: key %q not found", path, key)
			}
			cur = next
		default:
			return nil, fmt.Errorf("path %q: cannot descend into %T at %q", path, cur, segment)
		}
		if hasIdx {
			arr, ok := cur.([]any)
			if !ok {
				return nil, fmt.Errorf("path %q: %q is not an array", path, key)
			}
			if idx < 0 || idx >= len(arr) {
				return nil, fmt.Errorf("path %q: index %d out of range", path, idx)
			}
			cur = arr[idx]
		}
	}
	return cur, nil
}

func parseSegment(seg string) (key string, idx int, hasIdx bool) {
	open := strings.Index(seg, "[")
	if open < 0 {
		return seg, 0, false
	}
	close := strings.Index(seg, "]")
	if close < open+1 {
		return seg, 0, false
	}
	n, err := strconv.Atoi(seg[open+1 : close])
	if err != nil {
		return seg, 0, false
	}
	return seg[:open], n, true
}
