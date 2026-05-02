package server

// matches reports whether ev satisfies m. Empty matcher fields are wildcards;
// the literal value "any" is also treated as a wildcard.
func matches(m AchievementMatcher, ev Event) bool {
	if m.Actor != "" && m.Actor != "any" && m.Actor != ev.Actor {
		return false
	}
	if m.Action != "" && m.Action != "any" && m.Action != ev.Action {
		return false
	}
	if m.Target != "" && m.Target != "any" && m.Target != ev.Target {
		return false
	}
	if m.Result != "" && m.Result != "any" && m.Result != ev.Result {
		return false
	}
	if m.Key != "" && m.Key != "any" {
		evKey, _ := ev.Args["key"].(string)
		if m.Key != evKey {
			return false
		}
	}
	return true
}

// evalAchievements returns IDs newly unlocked by ev, given the current set.
func evalAchievements(stage *Stage, ev Event, current []string) []string {
	var unlocked []string
	for _, def := range stage.AchievementDefs {
		if has(current, def.ID) || has(unlocked, def.ID) {
			continue
		}
		if matches(def.When, ev) {
			unlocked = append(unlocked, def.ID)
		}
	}
	return unlocked
}
