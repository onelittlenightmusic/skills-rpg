package server

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Worlds.
//
// A world is a directory of stages under stages/, plus a world.yaml naming it.
// The dungeon is the exception and stays where it has always been — the flat
// stages/*.yaml — because worlds already exist under that layout and moving them
// would strand anyone's save.
//
// Progress is kept per world, in a state file of its own. That is what makes
// switching lossless: mywant's `world open` has to snapshot before switching
// because its worlds share one live canvas, and here they share nothing, so
// there is no save verb and nothing a switch can overwrite. The dungeon keeps
// current.yaml; every other world gets current-<world>.yaml.
//
// The alternative — one GameState holding every world — was rejected for what it
// would cost everywhere else: every read of CurrentStage in rules, goals and
// hooks would have had to learn which world it was in, to express something the
// filesystem already expresses.

// DungeonWorld is the world the flat stages/ directory is, and the world a save
// written before worlds existed belongs to.
const DungeonWorld = "dungeon"

// World is what a world.yaml says about itself.
type World struct {
	ID          string `yaml:"id" json:"id"`
	Title       string `yaml:"title" json:"title"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	FirstStage  string `yaml:"first_stage" json:"first_stage"`
	// UnlockedBy is an achievement from an earlier world. It gates the OFFER,
	// not the switch: an explicit switch always works, the same way debug/jump
	// does, because demoing or testing a world you have not earned is normal.
	UnlockedBy  string `yaml:"unlocked_by,omitempty" json:"unlocked_by,omitempty"`
	MywantWorld string `yaml:"mywant_world,omitempty" json:"mywant_world,omitempty"`

	Locales map[string]struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
	} `yaml:"locales,omitempty" json:"-"`
}

// stagesRoot is where this server reads stages from, whichever way it was
// configured — a directory on disk, or the embedded copy.
func (s *Server) stagesRoot() (fs.FS, string) {
	if s.cfg.StagesDir == "" {
		sub, _ := fs.Sub(DefaultStagesFS, "stages")
		return sub, "."
	}
	return os.DirFS(s.cfg.StagesDir), "."
}

// worldDir is where a world's stages live, relative to the stages root.
func worldDir(world string) string {
	if world == "" || world == DungeonWorld {
		return "."
	}
	return world
}

// listWorlds finds every world this build can play: the dungeon, plus each
// subdirectory of stages/ carrying a world.yaml.
func (s *Server) listWorlds() []World {
	root, base := s.stagesRoot()
	out := []World{{
		ID:         DungeonWorld,
		Title:      "The Dungeon",
		FirstStage: "stage1",
	}}
	entries, err := fs.ReadDir(root, base)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		w, err := readWorldFile(root, filepath.Join(base, e.Name(), "world.yaml"))
		if err != nil {
			continue // a directory without a world.yaml is not a world
		}
		if w.ID == "" {
			w.ID = e.Name()
		}
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func readWorldFile(f fs.FS, path string) (World, error) {
	var w World
	if err := readYAMLFromFS(f, path, &w); err != nil {
		return World{}, err
	}
	return w, nil
}

func (s *Server) findWorld(id string) (World, bool) {
	for _, w := range s.listWorlds() {
		if strings.EqualFold(w.ID, id) {
			return w, true
		}
	}
	return World{}, false
}

// statePathFor is where a world's progress is written. The dungeon keeps
// current.yaml so that a save written before worlds existed is still its save.
func (s *Server) statePathFor(world string) string {
	if world == "" || world == DungeonWorld {
		return filepath.Join(s.cfg.DataDir, "current.yaml")
	}
	return filepath.Join(s.cfg.DataDir, "current-"+world+".yaml")
}

// loadWorldLocked builds the in-memory state for a world: its stage definitions
// from YAML, and its progress from that world's own state file if it has one.
func (s *Server) loadWorldLocked(world string) error {
	root, base := s.stagesRoot()
	dir := filepath.Join(base, worldDir(world))

	fileStages, locales, _, err := loadStagesFromFS(root, dir)
	if err != nil {
		return fmt.Errorf("load stages for world %q: %w", world, err)
	}
	if len(fileStages) == 0 {
		return fmt.Errorf("world %q has no stages", world)
	}
	s.locales = locales
	s.pristine = fileStages

	var st GameState
	if err := readYAML(s.statePathFor(world), &st); err == nil {
		// Stages added to the YAML set after this state was persisted become
		// playable without resetting progress.
		for id, def := range fileStages {
			if _, ok := st.Stages[id]; !ok {
				st.Stages[id] = copyStage(def)
			}
		}
		s.state = &st
		s.world = world
		s.seedDoorSatisfiedLocked()
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	gs, _, err := initialStateFromFS(root, dir)
	if err != nil {
		return err
	}
	s.state = gs
	s.world = world
	s.seedDoorSatisfiedLocked()
	return s.persistLocked()
}

// SwitchWorld puts the game in another world, keeping the one being left exactly
// as it stands. `stage` overrides where to land, for jumping straight to a stage
// while testing; empty means resume where that world was left.
func (s *Server) SwitchWorld(world, stage string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	w, ok := s.findWorld(world)
	if !ok {
		return fmt.Errorf("no world %q", world)
	}
	if w.ID == s.world && stage == "" {
		return nil // already there; switching to where you are is not an error
	}

	// The world being left is written before anything else moves, so a failure
	// below cannot cost it.
	if err := s.persistLocked(); err != nil {
		return fmt.Errorf("save world %q before leaving: %w", s.world, err)
	}
	if err := s.loadWorldLocked(w.ID); err != nil {
		return err
	}
	if stage != "" {
		if _, ok := s.state.Stages[stage]; !ok {
			return fmt.Errorf("world %q has no stage %q", w.ID, stage)
		}
		s.state.CurrentStage = stage
		if st := s.state.Stages[stage]; st != nil {
			s.state.You.Position = st.InitialPosition
		}
		s.seedDoorSatisfiedLocked()
	}
	return s.persistLocked()
}

// ResetWorld rebuilds one world from its YAMLs and leaves the others alone.
func (s *Server) ResetWorld(world string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	w, ok := s.findWorld(world)
	if !ok {
		return fmt.Errorf("no world %q", world)
	}
	if err := os.Remove(s.statePathFor(w.ID)); err != nil && !os.IsNotExist(err) {
		return err
	}
	if w.ID != s.world {
		return nil // rebuilt on next entry; the world in play is untouched
	}
	return s.loadWorldLocked(w.ID)
}

// WorldStatus is one row of GET /api/v1/worlds.
type WorldStatus struct {
	World
	Current  bool `json:"current"`
	Unlocked bool `json:"unlocked"`
	Cleared  int  `json:"cleared"`
	Stages   int  `json:"stages"`
}

func (s *Server) Worlds() []WorldStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	held := map[string]bool{}
	for _, a := range s.state.Achievements {
		held[a] = true
	}

	out := make([]WorldStatus, 0, 4)
	for _, w := range s.listWorlds() {
		row := WorldStatus{World: w, Current: w.ID == s.world}
		row.Unlocked = w.UnlockedBy == "" || held[w.UnlockedBy]
		if row.Current {
			row.Stages = len(s.state.Stages)
			for _, st := range s.state.Stages {
				if st.ClearedWhen != "" && held[st.ClearedWhen] {
					row.Cleared++
				}
			}
		} else {
			// Another world's progress lives in its own file. Reading it is
			// cheap and the alternative is reporting zero for every world you
			// are not standing in, which makes the list useless.
			var other GameState
			if readYAML(s.statePathFor(w.ID), &other) == nil {
				done := map[string]bool{}
				for _, a := range other.Achievements {
					done[a] = true
				}
				row.Stages = len(other.Stages)
				for _, st := range other.Stages {
					if st.ClearedWhen != "" && done[st.ClearedWhen] {
						row.Cleared++
					}
				}
			}
		}
		out = append(out, row)
	}
	return out
}

// CurrentWorld is which world is in play.
func (s *Server) CurrentWorld() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.world
}
