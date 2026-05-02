package server

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

var slotRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)

type SaveSummary struct {
	Stage             string `yaml:"stage" json:"stage"`
	Position          string `yaml:"position" json:"position"`
	AchievementsCount int    `yaml:"achievements_count" json:"achievements_count"`
	NextGoal          string `yaml:"next_goal" json:"next_goal"`
}

type SaveMeta struct {
	Slot          string      `yaml:"slot" json:"slot"`
	Name          string      `yaml:"name,omitempty" json:"name,omitempty"`
	CreatedAt     string      `yaml:"created_at" json:"created_at"`
	UpdatedAt     string      `yaml:"updated_at" json:"updated_at"`
	PlaytimeSec   int         `yaml:"playtime_seconds" json:"playtime_seconds"`
	SchemaVersion int         `yaml:"schema_version" json:"schema_version"`
	Summary       SaveSummary `yaml:"summary" json:"summary"`
}

func validSlot(slot string) error {
	if !slotRegex.MatchString(slot) {
		return fmt.Errorf("invalid slot name %q (allowed: %s)", slot, slotRegex.String())
	}
	return nil
}

func saveDir(dataDir, slot string) string {
	return filepath.Join(dataDir, "saves", slot)
}

func summaryFor(state *GameState) SaveSummary {
	return SaveSummary{
		Stage:             state.CurrentStage,
		Position:          state.You.Position,
		AchievementsCount: len(state.Achievements),
		NextGoal:          state.NextGoal.Text,
	}
}

func writeSlot(dataDir, slot, name string, state *GameState) (SaveMeta, error) {
	if err := validSlot(slot); err != nil {
		return SaveMeta{}, err
	}
	dir := saveDir(dataDir, slot)
	now := time.Now().UTC().Format(time.RFC3339)
	var meta SaveMeta
	if err := readYAML(filepath.Join(dir, "meta.yaml"), &meta); err != nil {
		meta = SaveMeta{Slot: slot, CreatedAt: now}
	}
	if name != "" {
		meta.Name = name
	}
	meta.Slot = slot
	meta.UpdatedAt = now
	meta.PlaytimeSec = state.PlaytimeSec
	meta.SchemaVersion = SchemaVersion
	meta.Summary = summaryFor(state)
	if meta.CreatedAt == "" {
		meta.CreatedAt = now
	}
	if err := writeYAMLAtomic(filepath.Join(dir, "state.yaml"), state); err != nil {
		return SaveMeta{}, err
	}
	if err := writeYAMLAtomic(filepath.Join(dir, "meta.yaml"), &meta); err != nil {
		return SaveMeta{}, err
	}
	return meta, nil
}

func readSlot(dataDir, slot string) (*GameState, SaveMeta, error) {
	if err := validSlot(slot); err != nil {
		return nil, SaveMeta{}, err
	}
	dir := saveDir(dataDir, slot)
	var meta SaveMeta
	if err := readYAML(filepath.Join(dir, "meta.yaml"), &meta); err != nil {
		return nil, SaveMeta{}, err
	}
	var st GameState
	if err := readYAML(filepath.Join(dir, "state.yaml"), &st); err != nil {
		return nil, SaveMeta{}, err
	}
	return &st, meta, nil
}

func deleteSlot(dataDir, slot string) error {
	if err := validSlot(slot); err != nil {
		return err
	}
	return os.RemoveAll(saveDir(dataDir, slot))
}

func listSlots(dataDir string) ([]SaveMeta, error) {
	root := filepath.Join(dataDir, "saves")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []SaveMeta{}, nil
		}
		return nil, err
	}
	out := []SaveMeta{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		var meta SaveMeta
		if err := readYAML(filepath.Join(root, e.Name(), "meta.yaml"), &meta); err != nil {
			continue
		}
		out = append(out, meta)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}
