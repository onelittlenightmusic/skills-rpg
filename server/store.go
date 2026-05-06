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

// loadStagesFromDir reads every *.yaml file in dir (skipping *-jp.yaml locale files)
// and returns stages keyed by Stage.ID, locale data keyed by stageID→lang, and load order.
func loadStagesFromDir(dir string) (map[string]*Stage, map[string]map[string]*StageLocale, []string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, nil, err
	}
	out := map[string]*Stage{}
	locales := map[string]map[string]*StageLocale{}
	var order []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := filepath.Ext(name)
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		// Skip locale-only files (e.g. stage1-jp.yaml).
		base := name[:len(name)-len(ext)]
		if len(base) > 3 && base[len(base)-3:] == "-jp" {
			continue
		}
		var raw stageRaw
		if err := readYAML(filepath.Join(dir, name), &raw); err != nil {
			return nil, nil, nil, fmt.Errorf("read %s: %w", name, err)
		}
		if raw.ID == "" {
			return nil, nil, nil, fmt.Errorf("%s: stage missing id", name)
		}
		out[raw.ID] = &raw.Stage
		if len(raw.Locales) > 0 {
			locales[raw.ID] = raw.Locales
		}
		order = append(order, raw.ID)
	}
	return out, locales, order, nil
}

// initialStateFromStages builds a fresh GameState from a stages directory.
// The first stage in lexical filename order is selected as the starting stage.
func initialStateFromStages(dir string) (*GameState, map[string]map[string]*StageLocale, error) {
	stages, locales, order, err := loadStagesFromDir(dir)
	if err != nil {
		return nil, nil, err
	}
	if len(stages) == 0 {
		return nil, nil, fmt.Errorf("no stages found in %s", dir)
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
	gs.NextGoal = computeNextGoal(gs, nil)
	return gs, locales, nil
}
