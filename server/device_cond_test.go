package server

import (
	"encoding/json"
	"strings"
	"testing"
)

// Every condition kind has to be able to say what is outstanding.
//
// "off" is not an instruction, and this world's devices cannot be switched on by
// anybody in the game — the player puts the city into the state the device is
// reading, or the door stays shut forever. So an unsatisfied condition that says
// nothing is a stage nobody can finish, and that is what this test is for.
func TestEveryUnsatisfiedConditionSaysWhatToDo(t *testing.T) {
	b := newFakeBoard()
	withBoard(t, b)

	kinds := map[string]*Device{
		"named":      {Reads: &ThingCond{Subtype: "person", Value: "Lira"}},
		"any value":  {Reads: &ThingCond{Subtype: "level"}},
		"pinned":     {Reads: &ThingCond{Subtype: "person", Value: "Lira", Pinned: true}},
		"used by":    {ReadsUsed: &UsageCond{Subtype: "person", Value: "Lira", MinUsers: 3}},
		"group":      {ReadsGroup: &GroupCond{Name: "district_a", MinMembers: 8}},
		"want":       {ReadsWant: &WantCond{Type: "sluice", MinCount: 1}},
		"want gone":  {ReadsWant: &WantCond{Name: "drain", Absent: true}},
		"connection": {ReadsConn: &ConnCond{From: "supply", To: "sink"}},
	}

	for kind, dev := range kinds {
		if kind == "want gone" {
			continue // satisfied on an empty board, and tested below
		}
		if dev.IsOn() {
			t.Errorf("%s: running on an empty board", kind)
		}
		need := dev.Need()
		if need == "" {
			t.Errorf("%s: shut, and says nothing about what would open it", kind)
			continue
		}
		// An instruction, not a status line: it has to end in a full stop and
		// contain a verb the player can act on.
		if !strings.HasSuffix(need, ".") {
			t.Errorf("%s: need is not a sentence: %q", kind, need)
		}
		if len(dev.Checks()) == 0 {
			t.Errorf("%s: no checks for the card to draw", kind)
		}
	}
}

// The stage's own how-to rides along with the generated statement of what.
func TestNeedHintIsAppended(t *testing.T) {
	withBoard(t, newFakeBoard())
	dev := &Device{
		Reads:    &ThingCond{Subtype: "level"},
		NeedHint: "The sluice by the wall takes one.",
	}
	if !strings.Contains(dev.Need(), "The sluice by the wall takes one.") {
		t.Errorf("stage hint dropped: %q", dev.Need())
	}
}

// A satisfied device asks for nothing. A card that keeps telling the player to
// do something they have already done is worse than one that says nothing.
func TestSatisfiedConditionsAskForNothing(t *testing.T) {
	b := newFakeBoard()
	withBoard(t, b)
	b.named[fbKey("person", "Lira")] = true
	b.pinned[fbKey("person", "Lira")] = true
	b.used[fbKey("person", "Lira")] = 3
	b.groups["district_a"] = 8
	b.wants["sluice::::"] = 1
	b.roads["supply→sink"] = true

	for kind, dev := range map[string]*Device{
		"named":      {Reads: &ThingCond{Subtype: "person", Value: "Lira"}},
		"pinned":     {Reads: &ThingCond{Subtype: "person", Value: "Lira", Pinned: true}},
		"used by":    {ReadsUsed: &UsageCond{Subtype: "person", Value: "Lira", MinUsers: 3}},
		"group":      {ReadsGroup: &GroupCond{Name: "district_a", MinMembers: 8}},
		"want":       {ReadsWant: &WantCond{Type: "sluice", MinCount: 1}},
		"want gone":  {ReadsWant: &WantCond{Name: "drain", Absent: true}},
		"connection": {ReadsConn: &ConnCond{From: "supply", To: "sink"}},
	} {
		if !dev.IsOn() {
			t.Errorf("%s: not running with its condition met", kind)
		}
		if n := dev.Need(); n != "" {
			t.Errorf("%s: still asking for something: %q", kind, n)
		}
	}
}

// The pinning half is only asked for once there is something to pin. Telling
// somebody to pin a name the city has never heard of sends them looking for a
// button that is not there.
func TestPinIsAskedForInOrder(t *testing.T) {
	b := newFakeBoard()
	withBoard(t, b)
	dev := &Device{Reads: &ThingCond{Subtype: "person", Value: "Lira", Pinned: true}}

	if strings.Contains(dev.Need(), "Pin") {
		t.Errorf("asked for a pin before anything was named: %q", dev.Need())
	}
	b.named[fbKey("person", "Lira")] = true
	if !strings.Contains(dev.Need(), "Pin") {
		t.Errorf("named but not pinned, and the pin was not asked for: %q", dev.Need())
	}
}

// The card reads `on`, `checks` and `need` out of the state, so they have to be
// in the JSON — whatever the condition is.
func TestDeviceJSONCarriesTheAnswer(t *testing.T) {
	withBoard(t, newFakeBoard())
	body, err := json.Marshal(Device{Label: "Usage Meter",
		ReadsUsed: &UsageCond{Subtype: "person", Value: "Lira", MinUsers: 2}})
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		On     bool    `json:"on"`
		Checks []Check `json:"checks"`
		Need   string  `json:"need"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	if out.On {
		t.Error("reported running on an empty board")
	}
	if len(out.Checks) == 0 || out.Need == "" {
		t.Errorf("device JSON does not say what it wants: %s", body)
	}
}
