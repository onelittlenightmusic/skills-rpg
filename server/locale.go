package server

// StageLocale holds translated display text for one language.
// Loaded from the locales.<lang> block in each stage YAML.
// Only text fields are present — game logic fields (doors, devices, etc.) are not overridden.
type StageLocale struct {
	Title         string                    `yaml:"title"`
	Description   string                    `yaml:"description"`
	Waypoints     map[string]WaypointLocale `yaml:"waypoints"`
	Devices       map[string]DeviceLocale   `yaml:"devices"`
	NextGoalRules []GoalLocale              `yaml:"next_goal_rules"`
	Narrations    []NarrationLocale         `yaml:"narrations"`
}

// WaypointLocale overrides the waypoint label for a locale.
type WaypointLocale struct {
	Label string `yaml:"label"`
}

// DeviceLocale overrides the device label for a locale.
type DeviceLocale struct {
	Label string `yaml:"label"`
}

// GoalLocale overrides goal text/hint for one entry in next_goal_rules.
// Positionally corresponds to the matching GoalRule in the stage.
type GoalLocale struct {
	Goal Goal `yaml:"goal"`
}

// NarrationLocale overrides the narrative text for one entry in narrations.
// Positionally corresponds to the matching NarrationDef in the stage.
type NarrationLocale struct {
	Situation     string             `yaml:"situation"`
	Lore          string             `yaml:"lore"`
	OnSuccess     string             `yaml:"on_success"`
	Conversations []ConversationLine `yaml:"conversations"`
}

// stageRaw is used only during YAML loading to capture the locales block
// without polluting the GameState-serialized Stage struct.
type stageRaw struct {
	Stage   `yaml:",inline"`
	Locales map[string]*StageLocale `yaml:"locales,omitempty"`
}

// mergeNarrationLocale returns n with non-empty locale fields overlaid.
func mergeNarrationLocale(n Narration, loc *NarrationLocale) Narration {
	if loc == nil {
		return n
	}
	if loc.Situation != "" {
		n.Situation = loc.Situation
	}
	if loc.Lore != "" {
		n.Lore = loc.Lore
	}
	if loc.OnSuccess != "" {
		n.OnSuccess = loc.OnSuccess
	}
	if len(loc.Conversations) > 0 {
		n.Conversations = loc.Conversations
	}
	return n
}

// applyGoalLocale returns g with non-empty locale fields overlaid.
func applyGoalLocale(g Goal, loc *GoalLocale) Goal {
	if loc == nil {
		return g
	}
	if loc.Goal.Text != "" {
		g.Text = loc.Goal.Text
	}
	if loc.Goal.Hint != "" {
		g.Hint = loc.Goal.Hint
	}
	return g
}
