package server

const SchemaVersion = 1

type GameState struct {
	SchemaVersion int               `yaml:"schema_version" json:"schema_version"`
	CurrentStage  string            `yaml:"current_stage" json:"current_stage"`
	You           Player            `yaml:"you" json:"you"`
	Chap          Chap              `yaml:"chap" json:"chap"`
	Stages        map[string]*Stage `yaml:"stages" json:"stages"`
	Achievements  []string          `yaml:"achievements" json:"achievements"`
	NextGoal      Goal              `yaml:"next_goal" json:"next_goal"`
	PlaytimeSec   int               `yaml:"playtime_seconds" json:"playtime_seconds"`
}

type Player struct {
	Position  string   `yaml:"position" json:"position"`
	Inventory []string `yaml:"inventory,omitempty" json:"inventory,omitempty"`
}

type Chap struct {
	ActiveSkill string `yaml:"active_skill,omitempty" json:"active_skill,omitempty"`
}

type Stage struct {
	ID              string               `yaml:"id" json:"id"`
	Title           string               `yaml:"title" json:"title"`
	Description     string               `yaml:"description,omitempty" json:"description,omitempty"`
	InitialPosition string               `yaml:"initial_position" json:"initial_position"`
	Waypoints       map[string]*Waypoint `yaml:"waypoints" json:"waypoints"`
	Doors           map[string]*Door     `yaml:"doors,omitempty" json:"doors,omitempty"`
	Devices         map[string]*Device   `yaml:"devices,omitempty" json:"devices,omitempty"`
	Items           map[string]*Item     `yaml:"items,omitempty" json:"items,omitempty"`
	AchievementDefs []AchievementDef     `yaml:"achievement_defs" json:"achievement_defs"`
	NextGoalRules   []GoalRule           `yaml:"next_goal_rules" json:"next_goal_rules"`
	Narrations      []NarrationDef       `yaml:"narrations,omitempty" json:"narrations,omitempty"`
	ClearedWhen     string               `yaml:"cleared_when" json:"cleared_when"`
	NextStage       string               `yaml:"next_stage,omitempty" json:"next_stage,omitempty"`
}

// Narration is the rich, story-style supplement attached to a control result.
// All fields are optional; only those filled in stage YAML are emitted.
type Narration struct {
	Situation    string        `yaml:"situation,omitempty" json:"situation,omitempty"`
	WhyRejected  string        `yaml:"why_rejected,omitempty" json:"why_rejected,omitempty"`
	WhoIsChap    string        `yaml:"who_is_chap,omitempty" json:"who_is_chap,omitempty"`
	HowToProceed []ProceedStep `yaml:"how_to_proceed,omitempty" json:"how_to_proceed,omitempty"`
	Lore         string        `yaml:"lore,omitempty" json:"lore,omitempty"`
	OnSuccess    string        `yaml:"on_success,omitempty" json:"on_success,omitempty"`
}

type ProceedStep struct {
	Title string `yaml:"title" json:"title"`
	Body  string `yaml:"body" json:"body"`
}

type NarrationMatch struct {
	Actor        string `yaml:"actor,omitempty" json:"actor,omitempty"`
	Action       string `yaml:"action,omitempty" json:"action,omitempty"`
	Target       string `yaml:"target,omitempty" json:"target,omitempty"`
	TargetPrefix string `yaml:"target_prefix,omitempty" json:"target_prefix,omitempty"`
	Result       string `yaml:"result,omitempty" json:"result,omitempty"`
	Key          string `yaml:"key,omitempty" json:"key,omitempty"`
}

type NarrationDef struct {
	Match     NarrationMatch `yaml:"match" json:"match"`
	Narration Narration      `yaml:",inline" json:",inline"`
}

type Waypoint struct {
	Label    string   `yaml:"label" json:"label"`
	Adjacent []string `yaml:"adjacent,omitempty" json:"adjacent,omitempty"`
}

type Door struct {
	Between         [2]string `yaml:"between" json:"between"`
	Open            bool      `yaml:"open" json:"open"`
	Locked          bool      `yaml:"locked" json:"locked"`
	Key             string    `yaml:"key,omitempty" json:"key,omitempty"`
	RequiresDevice  string    `yaml:"requires_device,omitempty" json:"requires_device,omitempty"`
	BlockedByDevice string    `yaml:"blocked_by_device,omitempty" json:"blocked_by_device,omitempty"`
}

type Device struct {
	Label           string `yaml:"label" json:"label"`
	On              bool   `yaml:"on" json:"on"`
	BlockedByDevice string `yaml:"blocked_by_device,omitempty" json:"blocked_by_device,omitempty"`
}

type Item struct {
	At     string `yaml:"at,omitempty" json:"at,omitempty"`
	HeldBy string `yaml:"held_by,omitempty" json:"held_by,omitempty"`
}

type Goal struct {
	Text          string `yaml:"text" json:"text"`
	Hint          string `yaml:"hint,omitempty" json:"hint,omitempty"`
	RequiredSkill string `yaml:"required_skill,omitempty" json:"required_skill,omitempty"`
	Cleared       bool   `yaml:"cleared,omitempty" json:"cleared,omitempty"`
}

type AchievementDef struct {
	ID   string             `yaml:"id" json:"id"`
	When AchievementMatcher `yaml:"when" json:"when"`
}

// AchievementMatcher: empty fields = wildcard. Actor "any" also = wildcard.
type AchievementMatcher struct {
	Actor  string `yaml:"actor,omitempty" json:"actor,omitempty"`
	Action string `yaml:"action,omitempty" json:"action,omitempty"`
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
	Result string `yaml:"result,omitempty" json:"result,omitempty"`
	Key    string `yaml:"key,omitempty" json:"key,omitempty"`
}

type GoalRule struct {
	IfMissing string `yaml:"if_missing" json:"if_missing"`
	Goal      Goal   `yaml:"goal" json:"goal"`
}

type Event struct {
	Actor  string         `json:"actor"`
	Action string         `json:"action"`
	Target string         `json:"target"`
	Args   map[string]any `json:"args,omitempty"`
	Result string         `json:"result"`
	Reason string         `json:"reason,omitempty"`
}

type ControlInput struct {
	Actor  string         `json:"actor"`
	Action string         `json:"action"`
	Target string         `json:"target"`
	Args   map[string]any `json:"args,omitempty"`
}

type ControlResult struct {
	OK                   bool           `json:"ok"`
	Actor                string         `json:"actor"`
	Action               string         `json:"action"`
	Target               string         `json:"target"`
	Reason               string         `json:"reason,omitempty"`
	Narration            *Narration     `json:"narration,omitempty"`
	Changes              map[string]any `json:"changes,omitempty"`
	AchievementsUnlocked []string       `json:"achievements_unlocked,omitempty"`
	NextGoal             *Goal          `json:"next_goal,omitempty"`
}

func has(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
