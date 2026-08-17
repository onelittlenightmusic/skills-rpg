package server

import (
	"fmt"
	"strings"
)

// Action names recognised by /control.
const (
	ActionObserve    = "observe"
	ActionMove       = "move"
	ActionPickup     = "pickup"
	ActionOpen       = "open"       // unlock+open folded into one
	ActionActivate   = "activate"   // chap activates a device
	ActionDeactivate = "deactivate" // chap deactivates a device
	ActionState      = "state"      // chap checks device state
	ActionAdvance    = "advance"    // you advances to the next stage after clearing
	ActionReturn     = "return"     // you walks back to the stage that leads here

	// The fortress's own verbs. All three belong to you, not to chap: the
	// dungeon's division was that chap acts on the world and you only look at
	// it, and these do not cross that line. Naming a value and putting it on
	// the board are not interventions in the city — they are things the player
	// says about it, and saying them is the whole subject of world 2.
	ActionInspect   = "inspect"    // you reads an item, a door, or a want's card
	ActionNameThing = "name_thing" // you tells the city a value's name
	ActionPinThing  = "pin_thing"  // you puts that name on the board to stay
)

const (
	ActorYou  = "you"
	ActorChap = "chap"
)

// actorAllowed reports whether the given actor is permitted to perform action.
// A rejected action still produces an event so achievements like
// `attempted_self_unlock` can match against `result: rejected`.
func actorAllowed(actor, action string) bool {
	switch actor {
	case ActorYou:
		switch action {
		case ActionObserve, ActionMove, ActionPickup, ActionAdvance, ActionReturn,
			ActionInspect, ActionNameThing, ActionPinThing:
			return true
		}
	case ActorChap:
		switch action {
		case ActionObserve, ActionOpen, ActionActivate, ActionDeactivate, ActionState:
			return true
		}
	}
	return false
}

func validActor(actor string) bool { return actor == ActorYou || actor == ActorChap }

// applyControl mutates state in-place per a control request, returning the
// event that occurred and a result describing the outcome (including any
// achievements newly unlocked and the recomputed next_goal).
func applyControl(state *GameState, in ControlInput, locale *StageLocale) (Event, ControlResult) {
	res := ControlResult{Actor: in.Actor, Action: in.Action, Target: in.Target}
	ev := Event{Actor: in.Actor, Action: in.Action, Target: in.Target, Args: in.Args, Result: "rejected"}

	if !validActor(in.Actor) {
		res.Reason = fmt.Sprintf("unknown actor %q (must be 'you' or 'chap')", in.Actor)
		ev.Reason = res.Reason
		return ev, finishResult(state, ev, res, locale)
	}
	if in.Action == "" {
		res.Reason = "action is required"
		ev.Reason = res.Reason
		return ev, finishResult(state, ev, res, locale)
	}
	if !actorAllowed(in.Actor, in.Action) {
		res.Reason = fmt.Sprintf("%s cannot perform %q", in.Actor, in.Action)
		ev.Reason = res.Reason
		return ev, finishResult(state, ev, res, locale)
	}

	stage := state.Stages[state.CurrentStage]
	if stage == nil {
		res.Reason = "no current stage"
		ev.Reason = res.Reason
		return ev, finishResult(state, ev, res, locale)
	}

	var (
		changes map[string]any
		err     error
	)
	switch in.Action {
	case ActionObserve:
		// observe is also reachable via GET /observe; no state change.
		changes = nil
	case ActionMove:
		changes, err = doMove(state, stage, in)
	case ActionPickup:
		changes, err = doPickup(state, stage, in)
	case ActionOpen:
		changes, err = doOpen(state, stage, in)
	case ActionActivate:
		changes, err = doActivate(state, stage, in)
	case ActionDeactivate:
		changes, err = doDeactivate(state, stage, in)
	case ActionState:
		changes, err = doState(stage, in)
	case ActionAdvance:
		changes, err = doAdvance(state, stage)
	case ActionReturn:
		changes, err = doReturn(state, stage)
	case ActionInspect:
		changes, err = doInspect(state, stage, in)
	case ActionNameThing:
		changes, err = doNameThing(state, stage, in)
	case ActionPinThing:
		changes, err = doPinThing(state, stage, in)
	default:
		err = fmt.Errorf("unknown action %q", in.Action)
	}

	if err != nil {
		res.Reason = err.Error()
		ev.Reason = err.Error()
		return ev, finishResult(state, ev, res, locale)
	}

	ev.Result = "ok"
	res.OK = true
	res.Changes = changes
	return ev, finishResult(state, ev, res, locale)
}

func doState(stage *Stage, in ControlInput) (map[string]any, error) {
	dev, ok := stage.Devices[in.Target]
	if !ok {
		return nil, fmt.Errorf("no such device %q", in.Target)
	}
	return map[string]any{"on": dev.On}, nil
}

// finishResult records the event's achievements + recomputes next_goal,
// then attaches them and any matching narration to res.
// locale is optional; when non-nil its text fields are applied to the output.
func finishResult(state *GameState, ev Event, res ControlResult, locale *StageLocale) ControlResult {
	stage := state.Stages[state.CurrentStage]
	if stage != nil {
		newly := evalAchievements(stage, ev, state.Achievements)
		if len(newly) > 0 {
			state.Achievements = append(state.Achievements, newly...)
			res.AchievementsUnlocked = newly
		}
		if n := findNarration(stage, ev, state.Achievements, locale); n != nil {
			cp := *n
			res.Narration = &cp
		}
	}
	g := computeNextGoal(state, locale)
	state.NextGoal = g
	res.NextGoal = &g
	return res
}

// findNarration returns the first matching narration with locale overlay applied, or nil.
func findNarration(stage *Stage, ev Event, achievements []string, locale *StageLocale) *Narration {
	for i, nd := range stage.Narrations {
		if nd.Match.matches(ev, achievements) {
			n := nd.Narration
			if locale != nil && i < len(locale.Narrations) {
				n = mergeNarrationLocale(n, &locale.Narrations[i])
			}
			return &n
		}
	}
	return nil
}

func (m NarrationMatch) matches(ev Event, achievements []string) bool {
	if m.Actor != "" && m.Actor != "any" && m.Actor != ev.Actor {
		return false
	}
	if m.Action != "" && m.Action != "any" && m.Action != ev.Action {
		return false
	}
	if m.Result != "" && m.Result != "any" && m.Result != ev.Result {
		return false
	}
	if m.Target != "" && m.Target != ev.Target {
		return false
	}
	if m.TargetPrefix != "" && !strings.HasPrefix(ev.Target, m.TargetPrefix) {
		return false
	}
	if m.Key != "" {
		evKey, _ := ev.Args["key"].(string)
		if m.Key != evKey {
			return false
		}
	}
	for _, a := range m.RequiresAchievements {
		if !has(achievements, a) {
			return false
		}
	}
	for _, a := range m.MissingAchievements {
		if has(achievements, a) {
			return false
		}
	}
	return true
}

func doMove(state *GameState, stage *Stage, in ControlInput) (map[string]any, error) {
	dst := in.Target
	if dst == "" {
		return nil, fmt.Errorf("move requires target waypoint")
	}
	cur := state.You.Position
	if _, ok := stage.Waypoints[cur]; !ok {
		return nil, fmt.Errorf("current position %q not in stage", cur)
	}
	if _, ok := stage.Waypoints[dst]; !ok {
		return nil, fmt.Errorf("waypoint %q does not exist in current stage", dst)
	}
	// Try direct adjacency first.
	wpCur := stage.Waypoints[cur]
	if has(wpCur.Adjacent, dst) {
		if err := blockedByAnyDoor(stage, cur, dst); err != nil {
			return nil, err
		}
		state.You.Position = dst
		return map[string]any{"you.position": dst}, nil
	}
	// Fall back to BFS pathfinding through open waypoints.
	path := findPath(stage, cur, dst)
	if path == nil {
		return nil, fmt.Errorf("no open path from %q to %q", cur, dst)
	}
	state.You.Position = dst
	return map[string]any{"you.position": dst}, nil
}

// findPath returns a path from src to dst using BFS, only traversing edges
// that are not blocked by a locked/closed door. Returns nil if no path exists.
func findPath(stage *Stage, src, dst string) []string {
	type node struct {
		id   string
		path []string
	}
	visited := map[string]bool{src: true}
	queue := []node{{src, []string{src}}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range stage.Waypoints[cur.id].Adjacent {
			if visited[next] {
				continue
			}
			if blockedByAnyDoor(stage, cur.id, next) != nil {
				continue
			}
			p := append(append([]string{}, cur.path...), next)
			if next == dst {
				return p
			}
			visited[next] = true
			queue = append(queue, node{next, p})
		}
	}
	return nil
}

func doPickup(state *GameState, stage *Stage, in ControlInput) (map[string]any, error) {
	id := in.Target
	if id == "" {
		return nil, fmt.Errorf("pickup requires target item")
	}
	item, ok := stage.Items[id]
	if !ok {
		return nil, fmt.Errorf("item %q not in stage", id)
	}
	if item.HeldBy == ActorYou {
		return nil, fmt.Errorf("item %q is already held", id)
	}
	if item.At != state.You.Position {
		return nil, fmt.Errorf("item %q is not at your current position", id)
	}
	item.At = ""
	item.HeldBy = ActorYou
	state.You.Inventory = append(state.You.Inventory, id)
	return map[string]any{
		"you.inventory":      state.You.Inventory,
		"items." + id + ".held_by": ActorYou,
	}, nil
}

// chapInventory returns item IDs currently held by chap in the given stage.
func chapInventory(stage *Stage) []string {
	var inv []string
	for id, item := range stage.Items {
		if item.HeldBy == ActorChap {
			inv = append(inv, id)
		}
	}
	return inv
}

func doOpen(state *GameState, stage *Stage, in ControlInput) (map[string]any, error) {
	id := in.Target
	if id == "" {
		return nil, fmt.Errorf("open requires target door")
	}
	door, ok := stage.Doors[id]
	if !ok {
		return nil, fmt.Errorf("door %q not in stage", id)
	}
	if door.Open {
		// Idempotent: door already in desired state — treat as success so achievements fire.
		return map[string]any{
			"doors." + id + ".open":   true,
			"doors." + id + ".locked": false,
		}, nil
	}
	if door.Key != "" {
		if in.Actor == ActorChap {
			triedKey, _ := in.Args["key"].(string)
			if triedKey != "" {
				if !has(chapInventory(stage), triedKey) {
					return nil, fmt.Errorf("chap does not hold key %q", triedKey)
				}
				if triedKey != door.Key {
					return nil, fmt.Errorf("wrong key: %q does not open door %q", triedKey, id)
				}
			} else if !has(chapInventory(stage), door.Key) && !has(state.You.Inventory, door.Key) {
				return nil, fmt.Errorf("door %q requires key %q", id, door.Key)
			}
		} else {
			if !has(state.You.Inventory, door.Key) {
				return nil, fmt.Errorf("door %q requires key %q", id, door.Key)
			}
		}
	}
	if door.RequiresDevice != "" {
		dev, ok := stage.Devices[door.RequiresDevice]
		if !ok || !dev.On {
			return nil, fmt.Errorf("door %q requires device %q to be active", id, door.RequiresDevice)
		}
	}
	if door.BlockedByDevice != "" {
		dev, ok := stage.Devices[door.BlockedByDevice]
		if ok && dev.On {
			return nil, fmt.Errorf("door %q is blocked while device %q is active", id, door.BlockedByDevice)
		}
	}
	// The fortress's conditions: facts about the board rather than about this
	// stage. See rules_things.go.
	if err := thingGate(stage, id, door); err != nil {
		return nil, err
	}
	door.Open = true
	door.Locked = false
	return map[string]any{
		"doors." + id + ".open":   true,
		"doors." + id + ".locked": false,
	}, nil
}

func doActivate(state *GameState, stage *Stage, in ControlInput) (map[string]any, error) {
	id := in.Target
	if id == "" {
		return nil, fmt.Errorf("activate requires target device")
	}
	dev, ok := stage.Devices[id]
	if !ok {
		return nil, fmt.Errorf("device %q not in stage", id)
	}
	if dev.On {
		// Idempotent: device already in desired state — treat as success so achievements fire.
		return map[string]any{"devices." + id + ".on": true}, nil
	}
	if dev.BlockedByDevice != "" {
		blocker, ok := stage.Devices[dev.BlockedByDevice]
		if ok && blocker.On {
			return nil, fmt.Errorf("device %q cannot be activated while %q is active", id, dev.BlockedByDevice)
		}
	}
	dev.On = true
	return map[string]any{"devices." + id + ".on": true}, nil
}

func doDeactivate(state *GameState, stage *Stage, in ControlInput) (map[string]any, error) {
	id := in.Target
	if id == "" {
		return nil, fmt.Errorf("deactivate requires target device")
	}
	dev, ok := stage.Devices[id]
	if !ok {
		return nil, fmt.Errorf("device %q not in stage", id)
	}
	if !dev.On {
		// Idempotent: device already in desired state — treat as success so achievements fire.
		return map[string]any{"devices." + id + ".on": false}, nil
	}
	dev.On = false
	return map[string]any{"devices." + id + ".on": false}, nil
}

// blockedByAnyDoor returns an error if any closed door exists between a and b.
func doAdvance(state *GameState, stage *Stage) (map[string]any, error) {
	nextID := stage.NextStage
	next, ok := state.Stages[nextID]
	if !ok || next == nil {
		return nil, fmt.Errorf("no next stage to advance to")
	}
	// Clearing is what earns you a stage you have never seen. Walking back into
	// one you have already been in is not that — you left it standing, and the
	// door you came out of is still there. Without this, going back one stage
	// stranded you: the clear condition was met in a life the stage no longer
	// remembers, so the way forward refused and there was no way on.
	if _, beenThere := state.StageExit[nextID]; !beenThere {
		if stage.ClearedWhen == "" || !has(state.Achievements, stage.ClearedWhen) {
			return nil, fmt.Errorf("current stage is not cleared yet")
		}
	}
	// Where you were standing when you left, so walking back in later puts you
	// there rather than at that stage's entrance.
	rememberExit(state, stage.ID)
	state.CurrentStage = nextID
	state.EventHistory = nil
	if next.InitialPosition != "" {
		state.You.Position = next.InitialPosition
	}
	return map[string]any{"current_stage": nextID}, nil
}

// doReturn walks you back into the stage that leads to this one.
//
// The counterpart of doAdvance, and deliberately built out of the same pieces:
// change the current stage, forget the per-stage event log, place you. What it
// must NOT do is what DebugJumpStage does — restore the stage from its pristine
// YAML and empty your inventory and achievements. That is a debug restart, and
// using it to travel meant every trip out and back re-armed the alarms, re-shut
// the doors and forgot what you were carrying. Walking through a door is not a
// reason for a stage to forget you were ever in it.
//
// No "cleared" requirement, unlike advancing: you may leave a stage unfinished
// and come back to it, and the way back is exactly how you would.
func doReturn(state *GameState, stage *Stage) (map[string]any, error) {
	prevID := previousStageOf(state, stage.ID)
	if prevID == "" {
		return nil, fmt.Errorf("no previous stage to return to")
	}
	prev, ok := state.Stages[prevID]
	if !ok || prev == nil {
		return nil, fmt.Errorf("previous stage %q is not loaded", prevID)
	}
	rememberExit(state, stage.ID)
	state.CurrentStage = prevID
	state.EventHistory = nil
	if at := state.StageExit[prevID]; at != "" {
		state.You.Position = at
	} else if prev.InitialPosition != "" {
		state.You.Position = prev.InitialPosition
	}
	return map[string]any{"current_stage": prevID}, nil
}

// previousStageOf answers "who leads here", by inverting the next_stage links
// the stages already carry. Nothing stores the backward link: two directions
// recorded separately are two things that can disagree, and a stage graph that
// disagrees with itself would send you back somewhere you have never been.
func previousStageOf(state *GameState, id string) string {
	for prevID, st := range state.Stages {
		if st != nil && st.NextStage == id {
			return prevID
		}
	}
	return ""
}

// rememberExit records where you were standing in a stage as you leave it.
func rememberExit(state *GameState, stageID string) {
	if stageID == "" || state.You.Position == "" {
		return
	}
	if state.StageExit == nil {
		state.StageExit = map[string]string{}
	}
	state.StageExit[stageID] = state.You.Position
}

func blockedByAnyDoor(stage *Stage, a, b string) error {
	for id, d := range stage.Doors {
		if (d.Between[0] == a && d.Between[1] == b) || (d.Between[0] == b && d.Between[1] == a) {
			if !d.Open {
				return fmt.Errorf("door %q between %s and %s is not open", id, a, b)
			}
		}
	}
	return nil
}
