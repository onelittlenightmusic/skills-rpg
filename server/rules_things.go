package server

import "fmt"

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
	// The distinction is the subject of fortress2 and is not a detail.
	Pinned(subtype, value string) bool
}

// Read-only on purpose. The game asks mywant what the city knows and never
// writes on the player's behalf: naming a value and pinning it are mywant's own
// operations, done in mywant, which is the thing this world is here to teach.

// emptyBoard is the board when mywant is not configured. Every question answers
// "no" and every act fails with a reason the player can act on, rather than the
// stage silently behaving as though the city had been told something it has not.
type emptyBoard struct{}

func (emptyBoard) Named(string, string) bool  { return false }
func (emptyBoard) Pinned(string, string) bool { return false }

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

// thingGate is the door check for the two board conditions, shaped like the
// device checks in doOpen so a door reads the same whichever world it is in.
func thingGate(stage *Stage, id string, door *Door) error {
	if c := door.RequiresThingNamed; c != nil {
		v := c.Want(stage)
		if v == "" {
			return fmt.Errorf("door %q wants a named %s and the stage does not say which", id, c.Subtype)
		}
		if !currentBoard.Named(c.Subtype, v) {
			return fmt.Errorf("%s built this, and the city has no %s by that name — add one and try again", v, c.Subtype)
		}
	}
	if c := door.RequiresThingPinned; c != nil {
		v := c.Want(stage)
		if v == "" {
			return fmt.Errorf("door %q wants a pinned %s and the stage does not say which", id, c.Subtype)
		}
		if !currentBoard.Pinned(c.Subtype, v) {
			return fmt.Errorf("%s is on the board only while something uses it — pin it so it stays, then try again", v)
		}
	}
	return nil
}
