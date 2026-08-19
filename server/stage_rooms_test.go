package server

import "testing"

// The shorthand has to stand for exactly what it replaced, or the stages that
// use it are a second dialect rather than the same stages written shorter.
func TestRoomsExpandToTheLonghand(t *testing.T) {
	st := &Stage{
		ID: "t",
		Rooms: []Room{
			{ID: "a", Label: "A"},
			{ID: "b", Label: "B", Door: &RoomDoor{ID: "gate", Device: "reader"}},
			{ID: "c", Label: "C"},
		},
	}
	if err := ExpandRooms(st); err != nil {
		t.Fatal(err)
	}

	if st.InitialPosition != "a" {
		t.Errorf("initial position: %q", st.InitialPosition)
	}
	if got := st.Waypoints["b"].Adjacent; len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("middle room is not joined both ways: %v", got)
	}
	d := st.Doors["gate"]
	if d == nil || d.Between != [2]string{"a", "b"} || !d.Locked || d.RequiresDevice != "reader" {
		t.Errorf("door: %+v", d)
	}
	if st.ClearedWhen != RoomReached("c") {
		t.Errorf("cleared_when: %q", st.ClearedWhen)
	}

	defined := map[string]AchievementMatcher{}
	for _, ad := range st.AchievementDefs {
		defined[ad.ID] = ad.When
	}
	if _, ok := defined[DeviceRunning("reader")]; !ok {
		t.Error("no achievement for the reader coming alive — nothing would move the goal on")
	}
	for _, want := range []string{RoomReached("b"), RoomReached("c")} {
		if _, ok := defined[want]; !ok {
			t.Errorf("no achievement for entering %s", want)
		}
	}
	if _, ok := defined[RoomReached("a")]; ok {
		t.Error("awarded an achievement for standing where the stage starts")
	}
}

// Shorthand is for saving typing, not for overruling the author.
func TestHandWrittenPartsWin(t *testing.T) {
	st := &Stage{
		ID:              "t",
		InitialPosition: "b",
		ClearedWhen:     "something_else",
		Waypoints:       map[string]*Waypoint{"a": {Label: "mine", Adjacent: []string{"b", "c"}}},
		Doors:           map[string]*Door{"gate": {Between: [2]string{"a", "c"}, Open: true}},
		AchievementDefs: []AchievementDef{{ID: "something_else", When: AchievementMatcher{Action: "observe"}}},
		Rooms: []Room{
			{ID: "a", Label: "A"},
			{ID: "b", Label: "B", Door: &RoomDoor{ID: "gate", Device: "reader"}},
			{ID: "c", Label: "C"},
		},
	}
	if err := ExpandRooms(st); err != nil {
		t.Fatal(err)
	}
	if st.InitialPosition != "b" || st.ClearedWhen != "something_else" {
		t.Errorf("overwrote what the stage said: %q / %q", st.InitialPosition, st.ClearedWhen)
	}
	if st.Waypoints["a"].Label != "mine" {
		t.Error("overwrote a hand-written waypoint")
	}
	if !st.Doors["gate"].Open {
		t.Error("overwrote a hand-written door")
	}
	if n := len(st.AchievementDefs); n != 4 {
		t.Errorf("expected the hand-written one plus reader/b/c, got %d", n)
	}
}

// A door into the first room has nothing on the other side of it.
func TestDoorOnTheFirstRoomIsAnError(t *testing.T) {
	st := &Stage{ID: "t", Rooms: []Room{{ID: "a", Door: &RoomDoor{ID: "gate"}}}}
	if err := ExpandRooms(st); err == nil {
		t.Error("accepted a door into the first room")
	}
}
