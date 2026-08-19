package server

import (
	"fmt"
	"testing"
)

// A board that remembers, so a stage can be played through without a mywant
// running. Naming and pinning are the two states the fortress turns on, and
// keeping them apart here is the point of the fake: a value can be named and
// not pinned, which is exactly the position fortress3 puts the player in.
type fakeBoard struct {
	named  map[string]bool
	pinned map[string]bool
	fail   bool
}

func newFakeBoard() *fakeBoard {
	return &fakeBoard{named: map[string]bool{}, pinned: map[string]bool{}}
}

func fbKey(subtype, value string) string { return subtype + "::" + value }

func (b *fakeBoard) Named(s, v string) bool  { return b.named[fbKey(s, v)] }
func (b *fakeBoard) Pinned(s, v string) bool { return b.pinned[fbKey(s, v)] }
func (b *fakeBoard) Name(s, v string) (string, error) {
	if b.fail {
		return "", fmt.Errorf("no city")
	}
	b.named[fbKey(s, v)] = true
	return "thg-" + fbKey(s, v), nil
}
func (b *fakeBoard) Pin(s, v string) error {
	if b.fail {
		return fmt.Errorf("no city")
	}
	if !b.named[fbKey(s, v)] {
		return fmt.Errorf("nothing named %q to pin", v)
	}
	b.pinned[fbKey(s, v)] = true
	return nil
}

// unnames is the redaction crew's move, for the test that fortress3's door is
// not satisfied by a name that was only ever borrowed.
func (b *fakeBoard) unpin(s, v string) { delete(b.pinned, fbKey(s, v)) }

func withBoard(t *testing.T, b board) {
	t.Helper()
	prev := currentBoard
	setBoard(b)
	t.Cleanup(func() { currentBoard = prev })
}

// A minimal stage in the fortress's shape: a door that reads the board, and an
// item carrying the name it is waiting for.
func markStage(cond string) *Stage {
	d := &Door{Between: [2]string{"here", "there"}, Locked: true, RequiresDevice: "registry"}
	dev := &Device{Label: "City Registry", Reads: &ThingCond{Subtype: "person", Value: "Lira"}}
	if cond == "pinned" {
		dev.Reads.Pinned = true
	}
	return &Stage{
		ID:              "t",
		InitialPosition: "here",
		Waypoints: map[string]*Waypoint{
			"here":  {Adjacent: []string{"there"}},
			"there": {Adjacent: []string{"here"}},
		},
		Doors:   map[string]*Door{"gate": d},
		Devices: map[string]*Device{"registry": dev},
		Items: map[string]*Item{"lira_mark": {HeldBy: "you", Value: "Lira"}},
	}
}

func stateFor(st *Stage) *GameState {
	return &GameState{
		SchemaVersion: SchemaVersion,
		CurrentStage:  st.ID,
		You:           Player{Position: st.InitialPosition},
		Stages:        map[string]*Stage{st.ID: st},
	}
}

// The condition reads the value off the item, so a stage does not have to write
// fortress2: naming is the whole move. The door opens on the name existing —
// fortress3: naming is NOT enough. This is the distinction the whole stage is
// built on — a name that is only referred to can be taken away, so the door
// Pinning something the city has not been told about names it on the way.
// Taking the pin away shuts the door again — which is what a redaction crew
// amounts to from the door's point of view, and is worth pinning down so a
// fortress1: inspecting a door reports what it is waiting for, which is how the
// player finds out that the field is blank. chap cannot tell them.
func TestInspectDoorReportsWhatItWaitsFor(t *testing.T) {
	b := newFakeBoard()
	withBoard(t, b)
	st := markStage("named")
	state := stateFor(st)

	_, res := applyControl(state, ControlInput{Actor: ActorYou, Action: ActionInspect, Target: "gate"}, nil)
	if !res.OK {
		t.Fatalf("inspect failed: %s", res.Reason)
	}
	// A door reports the device powering it, and that device's state — which is
	// what a player standing at a dead gate needs to know.
	if res.Changes["requires_device"] != "registry" {
		t.Errorf("inspect did not name the device: %v", res.Changes)
	}
	if res.Changes["device_on"] != false {
		t.Errorf("inspect claimed the reader was running before anything was named: %v", res.Changes)
	}
}

func TestInspectItemReportsItsValue(t *testing.T) {
	withBoard(t, newFakeBoard())
	st := markStage("named")
	state := stateFor(st)

	_, res := applyControl(state, ControlInput{Actor: ActorYou, Action: ActionInspect, Target: "lira_mark"}, nil)
	if !res.OK {
		t.Fatalf("inspect failed: %s", res.Reason)
	}
	if res.Changes["value"] != "Lira" {
		t.Errorf("inspect did not read the mark: %v", res.Changes)
	}
}

// The three fortress verbs belong to you. chap acts on the world; naming a value
// Without a mywant the board says no and says why, rather than letting a stage