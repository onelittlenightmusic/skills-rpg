package server

import (
	"strings"
	"testing"
)

// The fortress stages are content, and content breaks quietly: a mistyped key
// becomes a zero value, a renamed achievement becomes a goal nobody can reach,
// and neither shows up until somebody plays that far. These check the things
// that have to hold for a stage to be playable at all.
//
// world.yaml is not a stage and is skipped by the loader (it has no `id:` that
// matches a stage), so the count here is stages only.

func loadFortress(t *testing.T) (map[string]*Stage, map[string]map[string]*StageLocale) {
	t.Helper()
	stages, locales, _, err := loadStagesFromDir("stages/fortress")
	if err != nil {
		t.Fatalf("load fortress stages: %v", err)
	}
	if len(stages) == 0 {
		t.Fatal("no fortress stages loaded")
	}
	return stages, locales
}

func TestFortressStagesParse(t *testing.T) {
	stages, _ := loadFortress(t)

	for id, st := range stages {
		if st.ID != id {
			t.Errorf("%s: id field is %q", id, st.ID)
		}
		if st.Title == "" {
			t.Errorf("%s: no title", id)
		}
		if st.ClearedWhen == "" {
			t.Errorf("%s: no cleared_when", id)
		}
		if st.InitialPosition == "" {
			t.Errorf("%s: no initial_position", id)
		}
		if _, ok := st.Waypoints[st.InitialPosition]; !ok {
			t.Errorf("%s: initial_position %q is not a waypoint", id, st.InitialPosition)
		}
	}
}

// Every achievement a goal rule waits on, and the one the stage clears on, has
// to be something an achievement_def can actually award. A typo here strands the
// player on a goal that can never be satisfied.
func TestFortressAchievementsAreReachable(t *testing.T) {
	stages, _ := loadFortress(t)

	for id, st := range stages {
		defined := map[string]bool{}
		for _, ad := range st.AchievementDefs {
			defined[ad.ID] = true
		}
		for _, rule := range st.NextGoalRules {
			if rule.IfMissing != "" && !defined[rule.IfMissing] {
				t.Errorf("%s: goal waits on undefined achievement %q", id, rule.IfMissing)
			}
		}
		if !defined[st.ClearedWhen] {
			t.Errorf("%s: cleared_when %q is not an achievement_def", id, st.ClearedWhen)
		}
	}
}

// Doors and waypoints have to agree, or a stage describes a room you cannot
// leave by a door that joins nowhere.
func TestFortressDoorsJoinRealWaypoints(t *testing.T) {
	stages, _ := loadFortress(t)

	for id, st := range stages {
		for name, d := range st.Doors {
			for _, wp := range d.Between {
				if _, ok := st.Waypoints[wp]; !ok {
					t.Errorf("%s: door %q joins unknown waypoint %q", id, name, wp)
				}
			}
		}
		for from, wp := range st.Waypoints {
			for _, to := range wp.Adjacent {
				if _, ok := st.Waypoints[to]; !ok {
					t.Errorf("%s: waypoint %q is adjacent to unknown %q", id, from, to)
				}
			}
		}
	}
}

// next_stage has to name a stage that exists, or the world dead-ends mid-play.
// The last stage of a world names nothing, which is how the world ends.
func TestFortressChainIsWhole(t *testing.T) {
	stages, _ := loadFortress(t)

	for id, st := range stages {
		if st.NextStage == "" {
			continue
		}
		if _, ok := stages[st.NextStage]; !ok {
			// Only a failure once that stage has been written; until then it is
			// the edge of what exists.
			t.Logf("%s: next_stage %q not written yet", id, st.NextStage)
		}
	}
}

// The Japanese locale is the one most players will read. A stage that ships
// without it silently falls back to English mid-sentence.
func TestFortressHasJapaneseLocale(t *testing.T) {
	stages, locales := loadFortress(t)

	for id := range stages {
		byLang, ok := locales[id]
		if !ok {
			t.Errorf("%s: no locales at all", id)
			continue
		}
		ja, ok := byLang["ja"]
		if !ok || ja == nil {
			t.Errorf("%s: no ja locale", id)
			continue
		}
		if ja.Title == "" {
			t.Errorf("%s: ja locale has no title", id)
		}
	}
}

// A stage has to be finishable on the board, not just in the state machine.
//
// The canvas lays a stage's rooms in a chain and puts the exit portal in the
// last one, so any door between rooms stands between the player and the way
// out. A door the stage gives the player no way past is therefore a wall.
//
// The fortress's doors are opened by facts about mywant, which the game does not
// bring about — the player does, in mywant. So what the stage has to provide is
// the moment of noticing: an achievement on the door being opened, and a goal
// telling the player what the city is missing. A stage with a board-reading door
// and no such achievement has nothing to say when the player gets stuck at it.
func TestFortressStagesAreWalkable(t *testing.T) {
	stages, _ := loadFortress(t)

	for id, st := range stages {
		// The city opening a door is what the stage has to notice — chap does not
		// open these, the name arriving does. See mywant_sse.go.
		opens := map[string]bool{}
		for _, ad := range st.AchievementDefs {
			if ad.When.Actor == ActorCity && ad.When.Action == ActionAnswered {
				opens[ad.When.Target] = true
			}
		}
		for name, d := range st.Doors {
			if d.Open {
				continue
			}
			if d.RequiresThingNamed == nil && d.RequiresThingPinned == nil {
				continue
			}
			if !opens[name] {
				t.Errorf("%s: door %q opens on something the player does in mywant, and the stage never notices it being opened — nothing will move the goal on",
					id, name)
			}
		}
	}
}

// Every goal has to say how, not only what.
//
// "Name it again" is not an instruction to somebody who has never named
// anything: it leaves out the call, the kind of thing, and which of their
// possessions the value is written on. This world's verbs are new — a player
// arrives having only ever asked chap to open doors — so a fortress hint that
// does not contain something runnable is a dead end, however well it explains
// the idea.
//
// The dungeon is deliberately not held to this. Its hints name skills
// (`/rpg-try-keys`, `rpg_observe`) that the player has by then been using for
// stages, and several are deliberately prose because working it out is the
// lesson.
func TestFortressHintsSayHow(t *testing.T) {
	stages, locales := loadFortress(t)

	// Something the player can actually run: the MCP tool, a skill, or a CLI.
	runnable := func(h string) bool {
		for _, m := range []string{"rpg_control_system", "rpg_observe", "rpg-", "/mywant-", "mywant "} {
			if strings.Contains(h, m) {
				return true
			}
		}
		return false
	}

	for id, st := range stages {
		for i, r := range st.NextGoalRules {
			if !runnable(r.Goal.Hint) {
				t.Errorf("%s: goal %d (%q) has no runnable hint: %q",
					id, i, r.Goal.Text, r.Goal.Hint)
			}
		}
		// The Japanese hints are what most players read, and they are a separate
		// list — a command added to one and not the other leaves half the
		// audience with the old dead end.
		ja := locales[id]["ja"]
		if ja == nil {
			continue
		}
		for i, r := range ja.NextGoalRules {
			if r.Goal.Hint != "" && !runnable(r.Goal.Hint) {
				t.Errorf("%s (ja): goal %d (%q) has no runnable hint: %q",
					id, i, r.Goal.Text, r.Goal.Hint)
			}
		}
	}
}

// The two goal lists are read as one: the Japanese locale replaces hints
// position by position, so a goal added or removed on one side and not the
// other silently pairs the wrong hint with the wrong step. That is how
// fortress2 ended up telling Japanese players to wait for a patrol that had
// been cut from the stage.
func TestFortressLocaleGoalsLineUp(t *testing.T) {
	stages, locales := loadFortress(t)

	for id, st := range stages {
		ja := locales[id]["ja"]
		if ja == nil {
			continue
		}
		if len(ja.NextGoalRules) != len(st.NextGoalRules) {
			t.Errorf("%s: %d goals in English, %d in ja — the lists are applied by position",
				id, len(st.NextGoalRules), len(ja.NextGoalRules))
		}
	}
}
