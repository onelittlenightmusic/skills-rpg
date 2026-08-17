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
	d := &Door{Between: [2]string{"here", "there"}, Locked: true}
	c := &ThingCond{Subtype: "maker", FromItem: "lira_mark"}
	switch cond {
	case "named":
		d.RequiresThingNamed = c
	case "pinned":
		d.RequiresThingPinned = c
	}
	return &Stage{
		ID:              "t",
		InitialPosition: "here",
		Waypoints: map[string]*Waypoint{
			"here":  {Adjacent: []string{"there"}},
			"there": {Adjacent: []string{"here"}},
		},
		Doors: map[string]*Door{"gate": d},
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
// the name twice and a goal hint does not have to spoil it.
func TestThingCondReadsValueFromItem(t *testing.T) {
	st := markStage("named")
	if got := st.Doors["gate"].RequiresThingNamed.Want(st); got != "Lira" {
		t.Fatalf("want value from item = %q, want %q", got, "Lira")
	}
}

// fortress2: naming is the whole move. The door opens on the name existing —
// nothing has to be pinned yet.
func TestNamedDoorOpensOnceNamed(t *testing.T) {
	b := newFakeBoard()
	withBoard(t, b)
	st := markStage("named")
	state := stateFor(st)

	_, res := applyControl(state, ControlInput{Actor: ActorChap, Action: ActionOpen, Target: "gate"}, nil)
	if res.OK {
		t.Fatal("gate opened before the city had been told the name")
	}

	_, res = applyControl(state, ControlInput{
		Actor: ActorYou, Action: ActionNameThing, Target: "lira_mark",
		Args: map[string]any{"subtype": "maker"},
	}, nil)
	if !res.OK {
		t.Fatalf("naming failed: %s", res.Reason)
	}

	_, res = applyControl(state, ControlInput{Actor: ActorChap, Action: ActionOpen, Target: "gate"}, nil)
	if !res.OK {
		t.Fatalf("gate stayed shut after naming: %s", res.Reason)
	}
}

// fortress3: naming is NOT enough. This is the distinction the whole stage is
// built on — a name that is only referred to can be taken away, so the door
// wants one that stands on its own.
func TestPinnedDoorIsNotSatisfiedByNamingAlone(t *testing.T) {
	b := newFakeBoard()
	withBoard(t, b)
	st := markStage("pinned")
	state := stateFor(st)

	_, res := applyControl(state, ControlInput{
		Actor: ActorYou, Action: ActionNameThing, Target: "lira_mark",
		Args: map[string]any{"subtype": "maker"},
	}, nil)
	if !res.OK {
		t.Fatalf("naming failed: %s", res.Reason)
	}

	_, res = applyControl(state, ControlInput{Actor: ActorChap, Action: ActionOpen, Target: "gate"}, nil)
	if res.OK {
		t.Fatal("a merely-named value satisfied a door that asks for a pinned one")
	}

	_, res = applyControl(state, ControlInput{
		Actor: ActorYou, Action: ActionPinThing, Target: "lira_mark",
		Args: map[string]any{"subtype": "maker"},
	}, nil)
	if !res.OK {
		t.Fatalf("pinning failed: %s", res.Reason)
	}

	_, res = applyControl(state, ControlInput{Actor: ActorChap, Action: ActionOpen, Target: "gate"}, nil)
	if !res.OK {
		t.Fatalf("gate stayed shut after pinning: %s", res.Reason)
	}
}

// Pinning something the city has not been told about names it on the way.
// Refusing would teach an ordering rule mywant does not have.
func TestPinNamesOnTheWay(t *testing.T) {
	b := newFakeBoard()
	withBoard(t, b)
	st := markStage("pinned")
	state := stateFor(st)

	_, res := applyControl(state, ControlInput{
		Actor: ActorYou, Action: ActionPinThing, Target: "lira_mark",
		Args: map[string]any{"subtype": "maker"},
	}, nil)
	if !res.OK {
		t.Fatalf("pin without a prior name failed: %s", res.Reason)
	}
	if !b.Named("maker", "Lira") {
		t.Error("pinning did not name it")
	}
}

// Taking the pin away shuts the door again — which is what a redaction crew
// amounts to from the door's point of view, and is worth pinning down so a
// later change cannot make the gate latch open.
func TestUnpinningShutsThePinnedDoor(t *testing.T) {
	b := newFakeBoard()
	withBoard(t, b)
	st := markStage("pinned")
	state := stateFor(st)

	applyControl(state, ControlInput{
		Actor: ActorYou, Action: ActionPinThing, Target: "lira_mark",
		Args: map[string]any{"subtype": "maker"},
	}, nil)
	b.unpin("maker", "Lira")

	st.Doors["gate"].Open = false
	st.Doors["gate"].Locked = true
	_, res := applyControl(state, ControlInput{Actor: ActorChap, Action: ActionOpen, Target: "gate"}, nil)
	if res.OK {
		t.Fatal("door opened for a value no longer on the board")
	}
}

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
	if res.Changes["subtype"] != "maker" {
		t.Errorf("inspect did not name the kind it wants: %v", res.Changes)
	}
	if res.Changes["satisfied"] != false {
		t.Errorf("inspect claimed satisfied before anything was named: %v", res.Changes)
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
// and putting it on the board are things the player says about it.
func TestFortressVerbsBelongToYou(t *testing.T) {
	for _, a := range []string{ActionInspect, ActionNameThing, ActionPinThing} {
		if !actorAllowed(ActorYou, a) {
			t.Errorf("you cannot %s", a)
		}
		if actorAllowed(ActorChap, a) {
			t.Errorf("chap should not be able to %s", a)
		}
	}
}

// Without a mywant the board says no and says why, rather than letting a stage
// behave as though a name had been given.
func TestNoMywantIsAnHonestRefusal(t *testing.T) {
	withBoard(t, emptyBoard{})
	st := markStage("named")
	state := stateFor(st)

	_, res := applyControl(state, ControlInput{
		Actor: ActorYou, Action: ActionNameThing, Target: "lira_mark",
		Args: map[string]any{"subtype": "maker"},
	}, nil)
	if res.OK {
		t.Fatal("naming succeeded with no city to tell")
	}
	if res.Reason == "" {
		t.Error("refused without saying why")
	}
}
