package server

import "encoding/json"

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
	EventHistory  []Event           `yaml:"event_history,omitempty" json:"event_history,omitempty"`
	// Where you were standing in each stage when you left it, so returning to
	// one puts you back where you were rather than at its entrance. Written on
	// the way out (advance and return), read on the way in.
	StageExit map[string]string `yaml:"stage_exit,omitempty" json:"stage_exit,omitempty"`
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
	MywantHooks     []MywantHook         `yaml:"mywant_hooks,omitempty" json:"mywant_hooks,omitempty"`
	MywantSetup     *MywantSetup         `yaml:"mywant_setup,omitempty" json:"mywant_setup,omitempty"`
}

// MywantSetup declares what the RPG server should create/configure in mywant
// when this stage is started or jumped to.
type MywantSetup struct {
	// CleanupLabel removes all existing mywant wants that carry every listed label.
	CleanupLabel map[string]string `yaml:"cleanup_label,omitempty" json:"cleanup_label,omitempty"`
	// RegisterWebhook makes the RPG server register itself as a lifecycle webhook
	// with mywant so it receives want_created events without any external script.
	RegisterWebhook bool `yaml:"register_webhook" json:"register_webhook"`
	// Wants lists the wants the RPG server creates in mywant on stage start.
	Wants []MywantSetupWant `yaml:"wants,omitempty" json:"wants,omitempty"`
	// OnStartRun is a list of shell commands executed after wants are created.
	OnStartRun []string `yaml:"on_start_run,omitempty" json:"on_start_run,omitempty"`
}

// MywantSetupWant describes a single want to create in mywant canvas.
type MywantSetupWant struct {
	Name   string            `yaml:"name" json:"name"`
	Type   string            `yaml:"type" json:"type"`
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Owner is the name of the parent want (ownerReference controller).
	Owner  string            `yaml:"owner,omitempty" json:"owner,omitempty"`
	Params map[string]any    `yaml:"params,omitempty" json:"params,omitempty"`
}

// MywantHook reacts to lifecycle events pushed from mywant (e.g. want_created).
// The RPG server registers its endpoint with mywant; mywant calls it on
// matching events; the server fires achievements and shell commands in response.
//
// Static hooks are defined in stage YAML (mywant_hooks:) for backwards compatibility.
// Dynamic hooks are registered via rpg_hook want type and arrive with Rule.Metadata
// in the payload — those bypass static matching entirely.
type MywantHook struct {
	Event                string            `yaml:"event" json:"event"`
	MatchName            string            `yaml:"match_name,omitempty" json:"match_name,omitempty"`
	MatchType            string            `yaml:"match_type,omitempty" json:"match_type,omitempty"`
	MatchLabel           map[string]string `yaml:"match_label,omitempty" json:"match_label,omitempty"`
	MatchOwner           string            `yaml:"match_owner,omitempty" json:"match_owner,omitempty"`
	MatchChildRole       string            `yaml:"match_child_role,omitempty" json:"match_child_role,omitempty"`
	SkipIfAchievement    string            `yaml:"skip_if_achievement,omitempty" json:"skip_if_achievement,omitempty"`
	RequireAchievements  []string          `yaml:"require_achievements,omitempty" json:"require_achievements,omitempty"`
	Activate             string            `yaml:"activate,omitempty" json:"activate,omitempty"`
	Run                  []string          `yaml:"run,omitempty" json:"run,omitempty"`
}

// MywantRuleRef is the rule info embedded in the lifecycle payload by mywant
// when a filtered rule (registered via rpg_hook want type) matches.
type MywantRuleRef struct {
	ID       string         `json:"id"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ConversationLine is a single line of dialogue spoken by a character.
type ConversationLine struct {
	Speaker string `yaml:"speaker" json:"speaker"`
	Text    string `yaml:"text" json:"text"`
}

// Narration is the rich, story-style supplement attached to a control result.
// All fields are optional; only those filled in stage YAML are emitted.
type Narration struct {
	Situation     string             `yaml:"situation,omitempty" json:"situation,omitempty"`
	WhyRejected   string             `yaml:"why_rejected,omitempty" json:"why_rejected,omitempty"`
	WhoIsChap     string             `yaml:"who_is_chap,omitempty" json:"who_is_chap,omitempty"`
	HowToProceed  []ProceedStep      `yaml:"how_to_proceed,omitempty" json:"how_to_proceed,omitempty"`
	Lore          string             `yaml:"lore,omitempty" json:"lore,omitempty"`
	OnSuccess     string             `yaml:"on_success,omitempty" json:"on_success,omitempty"`
	Conversations []ConversationLine `yaml:"conversations,omitempty" json:"conversations,omitempty"`
}

type ProceedStep struct {
	Title string `yaml:"title" json:"title"`
	Body  string `yaml:"body" json:"body"`
}

type NarrationMatch struct {
	Actor                string   `yaml:"actor,omitempty" json:"actor,omitempty"`
	Action               string   `yaml:"action,omitempty" json:"action,omitempty"`
	Target               string   `yaml:"target,omitempty" json:"target,omitempty"`
	TargetPrefix         string   `yaml:"target_prefix,omitempty" json:"target_prefix,omitempty"`
	Result               string   `yaml:"result,omitempty" json:"result,omitempty"`
	Key                  string   `yaml:"key,omitempty" json:"key,omitempty"`
	RequiresAchievements []string `yaml:"requires_achievements,omitempty" json:"requires_achievements,omitempty"`
	MissingAchievements  []string `yaml:"missing_achievements,omitempty" json:"missing_achievements,omitempty"`
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
	Between [2]string `yaml:"between" json:"between"`
	Open            bool      `yaml:"open" json:"open"`
	Locked          bool      `yaml:"locked" json:"locked"`
	Key             string    `yaml:"key,omitempty" json:"key,omitempty"`
	RequiresDevice  string    `yaml:"requires_device,omitempty" json:"requires_device,omitempty"`
	BlockedByDevice string    `yaml:"blocked_by_device,omitempty" json:"blocked_by_device,omitempty"`
}

// ThingCond names a value the world has to know about.
//
// The value is written here, plainly. An earlier version let a door defer to an
// item instead — `from_item: lira_mark` — so that the card would not give away
// a name the player was meant to go and read. What it gave away instead was an
// internal id: "name the person on lira_mark" means nothing to somebody seeing
// this for the first time, and no wording rescued it. A door that says which
// name is missing can be understood; a door that says which item holds it
// cannot.
type ThingCond struct {
	Subtype string `yaml:"subtype,omitempty" json:"subtype,omitempty"`
	Value   string `yaml:"value,omitempty" json:"value,omitempty"`
	// Pinned asks for the stronger fact: not merely known to the city, but
	// standing on the board because the player put it there.
	Pinned bool `yaml:"pinned,omitempty" json:"pinned,omitempty"`
}

// Want returns the value this condition is about.
func (c *ThingCond) Want(*Stage) string {
	if c == nil {
		return ""
	}
	return c.Value
}

type Device struct {
	Label string `yaml:"label" json:"label"`
	// On is the switch for devices chap operates. A device that reads the city
	// ignores it and answers from there instead — read IsOn(), never this.
	On              bool   `yaml:"on" json:"-"`
	BlockedByDevice string `yaml:"blocked_by_device,omitempty" json:"blocked_by_device,omitempty"`

	// Reads makes this device a reader of the city's registry: it is running
	// exactly while the value it names is known, or pinned. The door in front of
	// it says only `requires_device`, as every door in this game does.
	//
	// The condition lives here and not on the door because a door is a door.
	// World 1 established the shape — stage4's power_door requires a generator,
	// stage5's is blocked by an alarm — and a door that judged the state of the
	// city itself was a second, parallel mechanism for the same idea, with the
	// judgement in the one place nobody would look for it.
	Reads *ThingCond `yaml:"reads_thing,omitempty" json:"reads_thing,omitempty"`
}

// IsOn reports whether the device is running.
//
// A registry reader is running while the city holds what it is watching for;
// everything else keeps the switch chap flips.
func (d *Device) IsOn() bool {
	if d == nil {
		return false
	}
	if c := d.Reads; c != nil {
		if c.Pinned {
			return currentBoard.Pinned(c.Subtype, c.Value)
		}
		return currentBoard.Named(c.Subtype, c.Value)
	}
	return d.On
}

// MarshalJSON reports the device as it actually stands — clients read `on` out
// of the state, and a reader's answer has to be what they get.
//
// A reader also reports the two halves of its answer separately, because they
// are two different jobs for the player and "off" does not say which one is
// outstanding. A name the city has never heard of cannot be pinned; telling
// somebody to pin it is telling them to do the impossible in the wrong order.
func (d Device) MarshalJSON() ([]byte, error) {
	type plain Device
	out := struct {
		plain
		On     bool  `json:"on"`
		Named  *bool `json:"reads_named,omitempty"`
		Pinned *bool `json:"reads_is_pinned,omitempty"`
	}{plain: plain(d), On: d.IsOn()}
	if c := d.Reads; c != nil {
		named := currentBoard.Named(c.Subtype, c.Value)
		pinned := currentBoard.Pinned(c.Subtype, c.Value)
		out.Named, out.Pinned = &named, &pinned
	}
	return json.Marshal(out)
}

type Item struct {
	At     string `yaml:"at,omitempty" json:"at,omitempty"`
	HeldBy string `yaml:"held_by,omitempty" json:"held_by,omitempty"`
	// Value is what is written on it, for items that carry a name rather than
	// being one — a maker's mark struck with "Lira". Inspecting the item is how
	// the player reads it; naming it is how the city gets told.
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
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
	// Unless drops this rule once the named achievement is held, even though
	// IfMissing is still unmet.
	//
	// For steps that can stop being reachable. fortress1's opening beat is chap
	// trying the gate and nothing happening, which is only possible while the
	// city has no Lira — name her first and the refusal can never occur, so a
	// rule waiting for it would ask forever for something that cannot happen.
	// Skipping a beat is allowed; being told to reproduce it is not.
	Unless string `yaml:"unless,omitempty" json:"unless,omitempty"`
	Goal   Goal   `yaml:"goal" json:"goal"`
}

type Event struct {
	Actor     string         `json:"actor"`
	Action    string         `json:"action"`
	Target    string         `json:"target"`
	Args      map[string]any `json:"args,omitempty"`
	Result    string         `json:"result"`
	Reason    string         `json:"reason,omitempty"`
	Narration *Narration     `json:"narration,omitempty"`
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
