package server

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func writeYAMLAtomic(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	enc := yaml.NewEncoder(tmp)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		_ = enc.Close()
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := enc.Close(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

func readYAML(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, v)
}

// loadStagesFromDir reads every *.yaml file in dir as a Stage and returns them
// keyed by Stage.ID.
func loadStagesFromDir(dir string) (map[string]*Stage, []string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	out := map[string]*Stage{}
	var order []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".yaml" && filepath.Ext(name) != ".yml" {
			continue
		}
		var s Stage
		if err := readYAML(filepath.Join(dir, name), &s); err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", name, err)
		}
		if s.ID == "" {
			return nil, nil, fmt.Errorf("%s: stage missing id", name)
		}
		out[s.ID] = &s
		order = append(order, s.ID)
	}
	return out, order, nil
}

// initialStateFromStages builds a fresh GameState from a stages directory.
// The first stage in lexical filename order is selected as the starting stage.
func initialStateFromStages(dir string) (*GameState, error) {
	stages, order, err := loadStagesFromDir(dir)
	if err != nil {
		return nil, err
	}
	if len(stages) == 0 {
		return nil, fmt.Errorf("no stages found in %s", dir)
	}
	first := order[0]
	gs := &GameState{
		SchemaVersion: SchemaVersion,
		CurrentStage:  first,
		Stages:        stages,
		Achievements:  []string{},
	}
	if init := stages[first].InitialPosition; init != "" {
		gs.You.Position = init
	}
	gs.NextGoal = computeNextGoal(gs)
	return gs, nil
}
