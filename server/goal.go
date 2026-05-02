package server

// computeNextGoal walks the current stage's NextGoalRules in order and returns
// the first rule whose if_missing achievement is not yet unlocked. If the
// stage's cleared_when is satisfied, advances to next_stage when present.
//
// Mutates state.CurrentStage if a stage transition fires.
func computeNextGoal(state *GameState) Goal {
	for {
		stage, ok := state.Stages[state.CurrentStage]
		if !ok || stage == nil {
			return Goal{Text: "(no current stage)"}
		}
		if stage.ClearedWhen != "" && has(state.Achievements, stage.ClearedWhen) {
			if stage.NextStage != "" && state.Stages[stage.NextStage] != nil {
				state.CurrentStage = stage.NextStage
				if next := state.Stages[stage.NextStage]; next != nil && next.InitialPosition != "" {
					// advancing into a new stage repositions the player
					state.You.Position = next.InitialPosition
				}
				continue
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
}
