package server


// computeNextGoal returns the next goal for the current stage.
// Stage advancement is no longer automatic; the player must call action=advance.
func computeNextGoal(state *GameState) Goal {
	stage, ok := state.Stages[state.CurrentStage]
	if !ok || stage == nil {
		return Goal{Text: "(no current stage)"}
	}
	if stage.ClearedWhen != "" && has(state.Achievements, stage.ClearedWhen) {
		if stage.NextStage != "" && state.Stages[stage.NextStage] != nil {
			return Goal{
				Text: "🎉 ステージクリア！次のステージへ進もう",
				Hint: "rpg_control action=advance",
			}
		}
		return Goal{Text: "🎉 全ステージクリア", Cleared: true}
	}
	for _, rule := range stage.NextGoalRules {
		if !has(state.Achievements, rule.IfMissing) {
			return rule.Goal
		}
	}
	return Goal{Text: "ステージ進行中", Cleared: false}
}
