package server

import (
	"os"
	"path/filepath"
	"testing"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	s, err := NewServer(Config{
		DataDir:      dir,
		StagesDir:    "stages",
		SettingsPath: filepath.Join(dir, "settings.conf"),
	})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return s
}

// A game starts in the dungeon, and a save written before worlds existed is the
// dungeon's — which is why the dungeon keeps current.yaml rather than being
// moved to a file of its own.
func TestStartsInTheDungeonAndKeepsCurrentYaml(t *testing.T) {
	s := testServer(t)
	if got := s.CurrentWorld(); got != DungeonWorld {
		t.Fatalf("started in %q, want %q", got, DungeonWorld)
	}
	if _, err := os.Stat(filepath.Join(s.cfg.DataDir, "current.yaml")); err != nil {
		t.Errorf("dungeon did not write current.yaml: %v", err)
	}
}

func TestFortressIsAWorld(t *testing.T) {
	s := testServer(t)
	var found bool
	for _, w := range s.Worlds() {
		if w.ID == "fortress" {
			found = true
			if w.FirstStage != "fortress1" {
				t.Errorf("fortress first_stage = %q", w.FirstStage)
			}
			if w.MywantWorld == "" {
				t.Error("fortress does not name a mywant world")
			}
		}
	}
	if !found {
		t.Fatal("fortress is not listed as a world")
	}
}

// The point of the whole layer: leaving a world does not cost you anything in
// it. Each world writes to its own file, so there is nothing to save and nothing
// a switch can overwrite.
func TestSwitchingBackFindsTheWorldAsYouLeftIt(t *testing.T) {
	s := testServer(t)

	s.mu.Lock()
	s.state.CurrentStage = "stage4"
	s.state.Achievements = append(s.state.Achievements, "a-dungeon-thing")
	_ = s.persistLocked()
	s.mu.Unlock()

	if err := s.SwitchWorld("fortress", ""); err != nil {
		t.Fatalf("switch to fortress: %v", err)
	}
	if s.CurrentWorld() != "fortress" {
		t.Fatalf("did not switch, still in %q", s.CurrentWorld())
	}
	s.mu.Lock()
	if s.state.CurrentStage != "fortress1" {
		t.Errorf("fortress opened at %q, want fortress1", s.state.CurrentStage)
	}
	if has(s.state.Achievements, "a-dungeon-thing") {
		t.Error("the dungeon's achievements leaked into the fortress")
	}
	s.state.Achievements = append(s.state.Achievements, "a-fortress-thing")
	_ = s.persistLocked()
	s.mu.Unlock()

	if err := s.SwitchWorld(DungeonWorld, ""); err != nil {
		t.Fatalf("switch back: %v", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.CurrentStage != "stage4" {
		t.Errorf("dungeon resumed at %q, want stage4", s.state.CurrentStage)
	}
	if !has(s.state.Achievements, "a-dungeon-thing") {
		t.Error("the dungeon lost its progress")
	}
	if has(s.state.Achievements, "a-fortress-thing") {
		t.Error("the fortress's achievements leaked into the dungeon")
	}
}

func TestEachWorldHasItsOwnStateFile(t *testing.T) {
	s := testServer(t)
	if err := s.SwitchWorld("fortress", ""); err != nil {
		t.Fatalf("switch: %v", err)
	}
	if _, err := os.Stat(filepath.Join(s.cfg.DataDir, "current-fortress.yaml")); err != nil {
		t.Errorf("fortress has no state file of its own: %v", err)
	}
	if _, err := os.Stat(filepath.Join(s.cfg.DataDir, "current.yaml")); err != nil {
		t.Errorf("the dungeon's state file went missing: %v", err)
	}
}

// Resetting a world to retest a stage in it must not throw away another world.
func TestResetOneWorldLeavesTheOthers(t *testing.T) {
	s := testServer(t)
	s.mu.Lock()
	s.state.Achievements = append(s.state.Achievements, "a-dungeon-thing")
	_ = s.persistLocked()
	s.mu.Unlock()

	if err := s.SwitchWorld("fortress", ""); err != nil {
		t.Fatalf("switch: %v", err)
	}
	if err := s.ResetWorld("fortress"); err != nil {
		t.Fatalf("reset fortress: %v", err)
	}
	if err := s.SwitchWorld(DungeonWorld, ""); err != nil {
		t.Fatalf("switch back: %v", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !has(s.state.Achievements, "a-dungeon-thing") {
		t.Error("resetting the fortress reset the dungeon")
	}
}

// unlocked_by gates what is offered, not what is reachable. Demoing or testing a
// world nobody has earned is normal, and debug/jump has the same latitude.
func TestALockedWorldStillSwitches(t *testing.T) {
	s := testServer(t)
	var locked bool
	for _, w := range s.Worlds() {
		if w.ID == "fortress" {
			locked = !w.Unlocked
		}
	}
	if !locked {
		t.Skip("fortress is not gated in this build")
	}
	if err := s.SwitchWorld("fortress", ""); err != nil {
		t.Fatalf("a locked world refused an explicit switch: %v", err)
	}
}

func TestSwitchToAStageLandsThere(t *testing.T) {
	s := testServer(t)
	if err := s.SwitchWorld("fortress", "fortress3"); err != nil {
		t.Fatalf("switch: %v", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.CurrentStage != "fortress3" {
		t.Errorf("landed on %q, want fortress3", s.state.CurrentStage)
	}
	if s.state.You.Position != s.state.Stages["fortress3"].InitialPosition {
		t.Error("landed on the stage but not at its entrance")
	}
}

func TestUnknownWorldIsRefused(t *testing.T) {
	s := testServer(t)
	if err := s.SwitchWorld("atlantis", ""); err == nil {
		t.Fatal("switched to a world that does not exist")
	}
	if s.CurrentWorld() != DungeonWorld {
		t.Errorf("a failed switch moved us to %q", s.CurrentWorld())
	}
}
