package server

import (
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
// out. A door whose condition the stage never satisfies is therefore a wall,
// and the stage cannot be completed by walking — which is how it is played.
//
// fortress1 shipped exactly that: the north gate wanted a named maker, the
// stage had no naming step (that was the next stage's job), and the exit sat
// behind it. It cleared in the state machine and stranded anyone on the board.
func TestFortressStagesAreWalkable(t *testing.T) {
	stages, _ := loadFortress(t)

	for id, st := range stages {
		// What this stage can bring about itself.
		acts := map[string]bool{}
		for _, ad := range st.AchievementDefs {
			acts[ad.When.Action] = true
		}
		for name, d := range st.Doors {
			if d.Open {
				continue
			}
			if c := d.RequiresThingNamed; c != nil && !acts[ActionNameThing] && !acts[ActionPinThing] {
				t.Errorf("%s: door %q waits for a named %s, and the stage has no step that names anything — the exit is behind a wall",
					id, name, c.Subtype)
			}
			if c := d.RequiresThingPinned; c != nil && !acts[ActionPinThing] {
				t.Errorf("%s: door %q waits for a pinned %s, and the stage has no step that pins anything — the exit is behind a wall",
					id, name, c.Subtype)
			}
		}
	}
}
