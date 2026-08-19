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

// board is what the rules need to know about the city — the whole of it, and
// deliberately small. Every question here is one a player can answer by looking
// at the canvas, because every one of them is a stage's win condition.
type board interface {
	// CountNamed counts the values the city holds under this subtype. An empty
	// value counts them all, which is how a stage asks for "any level" without
	// dictating which.
	CountNamed(subtype, value string) int
	// Pinned reports whether a value is standing on the canvas by the user's own
	// decision, rather than merely being drawn because something refers to it.
	// The distinction is the subject of fortress2 and is not a detail.
	Pinned(subtype, value string) bool
	// UsedBy counts the live wants naming this value. A thing tile draws one
	// road per user, so this is the same number the player can see on the board
	// — which is the point of the stage that reads it.
	UsedBy(subtype, value string) int
	// GroupMembers counts what has been gathered under one name.
	GroupMembers(name string) int
	// CountWants counts wants by type, name and state; empty fields match
	// anything.
	CountWants(wantType, name, state string) int
	// WantState reads one field out of a want's current state — what it has
	// made, as against whether it is running.
	WantState(name, field string) string
	// Connected reports whether a road runs from one want to another, carrying
	// the named field (empty: any).
	Connected(from, to, field string) bool
	// ConnectionsFrom counts the roads leaving a want.
	ConnectionsFrom(name string) int
}

// Read-only on purpose. The game asks mywant what the city knows and never
// writes on the player's behalf: naming a value and pinning it are mywant's own
// operations, done in mywant, which is the thing this world is here to teach.

// emptyBoard is the board when mywant is not configured. Every question answers
// "no" and every act fails with a reason the player can act on, rather than the
// stage silently behaving as though the city had been told something it has not.
type emptyBoard struct{}

func (emptyBoard) CountNamed(string, string) int         { return 0 }
func (emptyBoard) Pinned(string, string) bool            { return false }
func (emptyBoard) UsedBy(string, string) int             { return 0 }
func (emptyBoard) GroupMembers(string) int               { return 0 }
func (emptyBoard) CountWants(string, string, string) int { return 0 }
func (emptyBoard) WantState(string, string) string       { return "" }
func (emptyBoard) Connected(string, string, string) bool { return false }
func (emptyBoard) ConnectionsFrom(string) int            { return 0 }

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
		if d.RequiresDevice != "" {
			out["requires_device"] = d.RequiresDevice
			if dev, ok := stage.Devices[d.RequiresDevice]; ok {
				out["device_on"] = dev.IsOn()
				out["waiting_for"] = dev.Label
			}
		}
		return out, nil
	}

	if dev, ok := stage.Devices[id]; ok {
		out := map[string]any{"device": id, "on": dev.IsOn(), "label": dev.Label}
		// What this one is watching the city for, and the one job outstanding —
		// the thing fortress1 asks the player to read off it. Whatever kind of
		// condition it turns out to be: the device answers, this only relays.
		if checks := dev.Checks(); len(checks) > 0 {
			out["checks"] = checks
		}
		if need := dev.Need(); need != "" {
			out["need"] = need
		}
		return out, nil
	}

	return nil, fmt.Errorf("nothing here called %q to look at", id)
}
