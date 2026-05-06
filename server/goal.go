package server

// computeNextGoal returns the next goal for the current stage.
// locale is optional; when non-nil its text fields override the English defaults.
func computeNextGoal(state *GameState, locale *StageLocale) Goal {
	stage, ok := state.Stages[state.CurrentStage]
	if !ok || stage == nil {
		return Goal{Text: "(no current stage)"}
	}
	if stage.ClearedWhen != "" && has(state.Achievements, stage.ClearedWhen) {
		if stage.NextStage != "" && state.Stages[stage.NextStage] != nil {
			return Goal{
				Text: "Stage cleared! Advance to the next stage",
				Hint: "rpg_control_system actor=you action=advance",
			}
		}
		return Goal{Text: "All stages cleared!", Cleared: true}
	}
	for i, rule := range stage.NextGoalRules {
		if !has(state.Achievements, rule.IfMissing) {
			g := rule.Goal
			if locale != nil && i < len(locale.NextGoalRules) {
				g = applyGoalLocale(g, &locale.NextGoalRules[i])
			}
			return g
		}
	}
	return Goal{Text: "Stage in progress", Cleared: false}
}
