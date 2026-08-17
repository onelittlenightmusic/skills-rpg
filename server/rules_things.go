package server

import (
	"fmt"
	"strings"
)

// The fortress's verbs, and the board they read.
//
// World 1's doors are opened by holding a key or by a device being on — facts
// the RPG server owns outright. World 2's are opened by facts about mywant: a
// value the city has been told the name of, and whether that name is standing on
// the board on the player's own say-so. The server does not own those, so it
// asks.
//
// It asks through a small interface rather than reaching for HTTP here. Rules
// stay a pure function of (state, input, board), which is what lets a stage be
// tested without a mywant running — and there is a lot of stage to test.

// board is what the rules need to know about the city's named values.
type board interface {
	// Named reports whether the city holds this value under this subtype.
	Named(subtype, value string) bool
	// Pinned reports whether it is standing on the canvas by the user's own
	// decision, rather than merely being drawn because something refers to it.
	// The distinction is the subject of fortress3 and is not a detail.
	Pinned(subtype, value string) bool
	// Name tells the city a value's name. Returns the thing's id.
	Name(subtype, value string) (string, error)
	// Pin puts an already-named value on the board to stay.
	Pin(subtype, value string) error
}

// emptyBoard is the board when mywant is not configured. Every question answers
// "no" and every act fails with a reason the player can act on, rather than the
// stage silently behaving as though the city had been told something it has not.
type emptyBoard struct{}

func (emptyBoard) Named(string, string) bool  { return false }
func (emptyBoard) Pinned(string, string) bool { return false }
func (emptyBoard) Name(string, string) (string, error) {
	return "", fmt.Errorf("this stage needs mywant, and none is configured (set MywantURL)")
}
func (emptyBoard) Pin(string, string) error {
	return fmt.Errorf("this stage needs mywant, and none is configured (set MywantURL)")
}

// currentBoard is swapped in by the server at startup. A package-level hook
// rather than a parameter threaded through applyControl, because every existing
// caller and every existing action would have had to grow an argument they have
// no use for.
var currentBoard board = emptyBoard{}

func setBoard(b board) {
	if b == nil {
		b = emptyBoard{}
	}
	currentBoard = b
}

// thingArgs pulls the subtype and value an action is about, from the request or
// from the item it names. `from_item` is how a stage avoids writing the value
// twice: the mark carries the name, and the goal text does not have to spoil it.
func thingArgs(stage *Stage, in ControlInput) (subtype, value string, err error) {
	subtype, _ = in.Args["subtype"].(string)
	value, _ = in.Args["value"].(string)

	item := in.Target
	if item == "" {
		item, _ = in.Args["from_item"].(string)
	}
	if value == "" && item != "" && stage != nil {
		if it, ok := stage.Items[item]; ok {
			value = it.Value
		}
	}

	subtype = strings.TrimSpace(subtype)
	value = strings.TrimSpace(value)
	if subtype == "" {
		return "", "", fmt.Errorf("name_thing requires a subtype (what kind of thing this is)")
	}
	if value == "" {
		return "", "", fmt.Errorf("name_thing requires a value, or an item that carries one")
	}
	return subtype, value, nil
}

// doInspect — you reads something: an item in hand, or a door standing here.
//
// It changes nothing, which is the point. The dungeon's split was that chap acts
// and you look; inspect is the looking made into a move you can be asked to
// make, so a stage can be cleared by having understood something rather than by
// having done something to it.
func doInspect(state *GameState, stage *Stage, in ControlInput) (map[string]any, error) {
	id := in.Target
	if id == "" {
		return nil, fmt.Errorf("inspect requires a target")
	}

	if it, ok := stage.Items[id]; ok {
		out := map[string]any{"item": id}
		if it.Value != "" {
			out["value"] = it.Value
		}
		if it.HeldBy != "" {
			out["held_by"] = it.HeldBy
		}
		if it.At != "" {
			out["at"] = it.At
		}
		return out, nil
	}

	if d, ok := stage.Doors[id]; ok {
		out := map[string]any{
			"door":   id,
			"open":   d.Open,
			"locked": d.Locked,
		}
		// What this door is waiting on, in the door's own terms. fortress1 is
		// cleared by reading exactly this and finding the field blank.
		if c := d.RequiresThingNamed; c != nil {
			out["waiting_for"] = "a named " + c.Subtype
			out["subtype"] = c.Subtype
			out["satisfied"] = currentBoard.Named(c.Subtype, c.Want(stage))
		}
		if c := d.RequiresThingPinned; c != nil {
			out["waiting_for"] = "a pinned " + c.Subtype
			out["subtype"] = c.Subtype
			out["satisfied"] = currentBoard.Pinned(c.Subtype, c.Want(stage))
		}
		return out, nil
	}

	if dev, ok := stage.Devices[id]; ok {
		return map[string]any{"device": id, "on": dev.On, "label": dev.Label}, nil
	}

	return nil, fmt.Errorf("nothing here called %q to look at", id)
}

// doNameThing — you tells the city a value's name.
func doNameThing(state *GameState, stage *Stage, in ControlInput) (map[string]any, error) {
	subtype, value, err := thingArgs(stage, in)
	if err != nil {
		return nil, err
	}
	id, err := currentBoard.Name(subtype, value)
	if err != nil {
		return nil, err
	}
	return map[string]any{"thing": id, "subtype": subtype, "value": value}, nil
}

// doPinThing — you puts a name on the board, to stay.
//
// Naming first is not enforced as a separate step: pinning something the city
// has not been told about names it on the way, because refusing would teach the
// player an ordering rule that does not exist in mywant.
func doPinThing(state *GameState, stage *Stage, in ControlInput) (map[string]any, error) {
	subtype, value, err := thingArgs(stage, in)
	if err != nil {
		return nil, err
	}
	if !currentBoard.Named(subtype, value) {
		if _, err := currentBoard.Name(subtype, value); err != nil {
			return nil, err
		}
	}
	if err := currentBoard.Pin(subtype, value); err != nil {
		return nil, err
	}
	return map[string]any{"pinned": value, "subtype": subtype}, nil
}

// thingGate is the door check for the two board conditions, shaped like the
// device checks in doOpen so a door reads the same whichever world it is in.
func thingGate(stage *Stage, id string, door *Door) error {
	if c := door.RequiresThingNamed; c != nil {
		v := c.Want(stage)
		if v == "" {
			return fmt.Errorf("door %q wants a named %s and the stage does not say which", id, c.Subtype)
		}
		if !currentBoard.Named(c.Subtype, v) {
			return fmt.Errorf("door %q is waiting for a %s the city has not been told the name of", id, c.Subtype)
		}
	}
	if c := door.RequiresThingPinned; c != nil {
		v := c.Want(stage)
		if v == "" {
			return fmt.Errorf("door %q wants a pinned %s and the stage does not say which", id, c.Subtype)
		}
		if !currentBoard.Pinned(c.Subtype, v) {
			return fmt.Errorf("door %q is waiting for a %s that stays on the board — naming it is not enough here", id, c.Subtype)
		}
	}
	return nil
}
