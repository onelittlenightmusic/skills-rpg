package server

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

//go:embed stages/*.yaml skills/*
var DefaultDataFS embed.FS

// DefaultStagesFS is a convenience alias for the embedded stages.
var DefaultStagesFS = DefaultDataFS

type Config struct {
	DataDir      string // ~/.mywant-rpg
	StagesDir    string // path to stages/ for bootstrap (empty means use embedded)
	Port         int    // 7100
	SettingsPath string // default: ~/.skills-rpg.conf
}

type Server struct {
	cfg          Config
	mu           sync.Mutex
	state        *GameState
	settings     Settings
	settingsPath string
	locales      map[string]map[string]*StageLocale // stageID → lang → locale
}

func NewServer(cfg Config) (*Server, error) {
	settingsPath := cfg.SettingsPath
	if settingsPath == "" {
		settingsPath = defaultSettingsPath()
	}
	settings, err := loadSettings(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}

	s := &Server{cfg: cfg, settingsPath: settingsPath, settings: settings}
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

	// Always load locale data fresh from stage YAMLs (not stored in current.yaml).
	var locales map[string]map[string]*StageLocale
	var err error

	if s.cfg.StagesDir == "" {
		embedded, _ := fs.Sub(DefaultStagesFS, "stages")
		_, locales, _, err = loadStagesFromFS(embedded, ".")
	} else {
		_, locales, _, err = loadStagesFromDir(s.cfg.StagesDir)
	}
	if err != nil {
		return fmt.Errorf("load locales: %w", err)
	}
	s.locales = locales

	var st GameState
	if err := readYAML(s.currentPath(), &st); err == nil {
		s.state = &st
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	var gs *GameState
	if s.cfg.StagesDir == "" {
		embedded, _ := fs.Sub(DefaultStagesFS, "stages")
		gs, _, err = initialStateFromFS(embedded, ".")
	} else {
		gs, _, err = initialStateFromStages(s.cfg.StagesDir)
	}
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

	var gs *GameState
	var locales map[string]map[string]*StageLocale
	var err error

	if s.cfg.StagesDir == "" {
		embedded, _ := fs.Sub(DefaultStagesFS, "stages")
		gs, locales, err = initialStateFromFS(embedded, ".")
	} else {
		gs, locales, err = initialStateFromStages(s.cfg.StagesDir)
	}

	if err != nil {
		return err
	}
	s.state = gs
	s.locales = locales
	return s.persistLocked()
}

func (s *Server) persistLocked() error {
	return writeYAMLAtomic(s.currentPath(), s.state)
}

// currentLocaleLocked returns the StageLocale for the current stage and language.
// Must be called with s.mu held.
func (s *Server) currentLocaleLocked() *StageLocale {
	lang := s.settings.Language
	if lang == "" || lang == "en" {
		return nil
	}
	langs := s.locales[s.state.CurrentStage]
	if langs == nil {
		return nil
	}
	return langs[lang]
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

	locale := s.currentLocaleLocked()
	ev := Event{Actor: actor, Action: ActionObserve, Target: target, Result: "ok"}
	res := finishResult(s.state, ev, ControlResult{
		OK: true, Actor: actor, Action: ActionObserve, Target: target,
	}, locale)
	ev.Narration = res.Narration

	// Compute hash of game-relevant state (excludes event_history to avoid self-invalidation).
	stateHash := gameStateHash(s.state)
	ev.Args = map[string]any{"state_hash": stateHash}

	// Skip appending if the last observe has the same state hash,
	// regardless of actor (chap/you are interchangeable for polling).
	last := len(s.state.EventHistory) - 1
	isDup := last >= 0 &&
		s.state.EventHistory[last].Action == ActionObserve &&
		s.state.EventHistory[last].Target == target &&
		s.state.EventHistory[last].Args != nil &&
		s.state.EventHistory[last].Args["state_hash"] == stateHash
	if !isDup {
		s.state.EventHistory = append(s.state.EventHistory, ev)
		const maxHistory = 20
		if len(s.state.EventHistory) > maxHistory {
			s.state.EventHistory = s.state.EventHistory[len(s.state.EventHistory)-maxHistory:]
		}
		if err := s.persistLocked(); err != nil {
			return nil, nil, err
		}
	}
	return subtree, &res, nil
}

// trimInactiveStages removes all stages except the current one from the
// full-state subtree so LLMs don't get confused by 9 stages of data (~90KB).
// A no-op when subtree is not the full state map.
func trimInactiveStages(subtree any, currentStage string) any {
	m, ok := subtree.(map[string]any)
	if !ok {
		return subtree
	}
	stages, ok := m["stages"].(map[string]any)
	if !ok {
		return subtree
	}
	cur := stages[currentStage]
	trimmed := make(map[string]any, len(m))
	for k, v := range m {
		trimmed[k] = v
	}
	trimmed["stages"] = map[string]any{currentStage: cur}
	return trimmed
}

// Control applies a control input.
func (s *Server) Control(in ControlInput) (ControlResult, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	locale := s.currentLocaleLocked()
	ev, res := applyControl(s.state, in, locale)
	ev.Narration = res.Narration
	s.state.EventHistory = append(s.state.EventHistory, ev)
	const maxHistory = 20
	if len(s.state.EventHistory) > maxHistory {
		s.state.EventHistory = s.state.EventHistory[len(s.state.EventHistory)-maxHistory:]
	}
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
	return computeNextGoal(s.state, s.currentLocaleLocked())
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
	s.state.EventHistory = nil
	s.state.NextGoal = computeNextGoal(s.state, s.currentLocaleLocked())
	return s.persistLocked()
}

// --- settings ---

func (s *Server) GetSettings() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.settings
}

func (s *Server) UpdateSettings(settings Settings) error {
	settings.setDefaults()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = settings
	return saveSettings(s.settingsPath, settings)
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
	g := computeNextGoal(s.state, s.currentLocaleLocked())
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

// gameStateHash hashes only game-world fields (position, doors, devices, achievements),
// deliberately excluding EventHistory to avoid self-invalidation on each poll.
func gameStateHash(gs *GameState) string {
	type doorSnap struct{ Open, Locked bool }
	type deviceSnap struct{ On bool }
	type snap struct {
		Stage        string
		Position     string
		Inventory    []string
		Achievements []string
		Doors        map[string]doorSnap
		Devices      map[string]deviceSnap
	}
	s := snap{
		Stage:        gs.CurrentStage,
		Position:     gs.You.Position,
		Inventory:    gs.You.Inventory,
		Achievements: gs.Achievements,
		Doors:        map[string]doorSnap{},
		Devices:      map[string]deviceSnap{},
	}
	if stage, ok := gs.Stages[gs.CurrentStage]; ok {
		for id, d := range stage.Doors {
			s.Doors[id] = doorSnap{d.Open, d.Locked}
		}
		for id, d := range stage.Devices {
			s.Devices[id] = deviceSnap{d.On}
		}
	}
	b, _ := json.Marshal(s)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}
