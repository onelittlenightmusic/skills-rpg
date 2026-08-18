package server

import (
	"bufio"
	"log"
	"net/http"
	"strings"
	"time"
)

// Watching mywant.
//
// The fortress's doors are opened by facts about mywant, and until now the game
// only ever learned those by being asked — a door checked the city at the moment
// somebody tried to open it. That is enough to unlock a door and not enough to
// say anything about it. The moment worth narrating is when the player names a
// value and a tile appears on the board with a road running to the gate; a game
// that finds out later, because you happened to poke it, cannot tell you that.
//
// mywant already says so out loud: /api/v1/events streams thing_changed and
// want_changed. So the game listens.
//
// What it does with them is deliberately narrow. It does not open doors, move
// anyone, or clear a stage — those stay the player's and chap's. It records that
// the city answered, as an ordinary event, which is all narration and
// achievements have ever needed to fire.

// ActorCity is the event actor for something the city did on its own — a value
// named, a name removed. Neither `you` nor `chap` did it, and pretending
// otherwise would put words in somebody's mouth.
const ActorCity = "city"

// ActionAnswered is recorded when a door's board condition, previously unmet,
// is met. Named for what the player sees: the gate answering its maker.
const ActionAnswered = "answered"

// ActionForgot is its opposite, and the reason this watches for changes in both
// directions rather than latching. A name can leave the city — deleted, or
// unpinned while nothing refers to it — and that is not a non-event: it is the
// Empire's whole method, performed on a name the player restored. A door that
// silently shut again would be the game declining to mention the one thing
// fortress2 is about.
const ActionForgot = "forgot"

// watchMywant subscribes to mywant's event stream for as long as the server
// runs, reconnecting when it drops. mywant restarting is normal and not worth a
// stack trace.
func (s *Server) watchMywant() {
	if s.cfg.MywantURL == "" {
		return
	}
	for {
		if err := s.streamMywantEvents(); err != nil {
			// Debug rather than error: not being able to reach mywant is the
			// ordinary state of a game played without one.
			log.Printf("[mywant-sse] %v; retrying", err)
		}
		time.Sleep(3 * time.Second)
	}
}

func (s *Server) streamMywantEvents() error {
	// No timeout: this request is meant to stay open. The package-level client
	// has one, which would cut the stream every ten seconds.
	client := &http.Client{}
	resp, err := client.Get(s.cfg.MywantURL + "/api/v1/events")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var eventType string
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "event:"):
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			// The payload is not read. Which thing changed does not matter —
			// what matters is whether any door in the stage the player is
			// standing in is now satisfied, and that is a question about the
			// board as a whole, asked below. Reading the id and trusting it
			// would also be wrong: a value can be named by something other than
			// the event that named it.
			if eventType == "thing_changed" {
				s.reconcileBoardDoors()
			}
			eventType = ""
		}
	}
	return sc.Err()
}

// reconcileBoardDoors records an event for each door in the current stage whose
// board condition has changed since last look — met, or no longer met.
//
// Both directions, and remembered per door rather than latched, because a name
// can come and go: pinned, unpinned, deleted, named again. Latching would have
// said "the gate answered" once and then narrated nothing for the rest of the
// world, including the beat fortress2 exists for.
func (s *Server) reconcileBoardDoors() {
	s.mu.Lock()
	stage := s.state.Stages[s.state.CurrentStage]
	if stage == nil {
		s.mu.Unlock()
		return
	}
	type change struct {
		door, value string
		met         bool
	}
	var news []change
	for id, d := range stage.Doors {
		c := d.RequiresThingNamed
		pinned := false
		if c == nil {
			c, pinned = d.RequiresThingPinned, true
		}
		if c == nil {
			continue
		}
		v := c.Want(stage)
		met := currentBoard.Named(c.Subtype, v)
		if pinned {
			met = currentBoard.Pinned(c.Subtype, v)
		}
		if was, seen := s.doorSatisfied[id]; seen && was == met {
			continue
		}
		if s.doorSatisfied == nil {
			s.doorSatisfied = map[string]bool{}
		}
		s.doorSatisfied[id] = met
		news = append(news, change{door: id, value: v, met: met})
	}
	if len(news) == 0 {
		s.mu.Unlock()
		return
	}
	locale := s.currentLocaleLocked()
	for _, ch := range news {
		action := ActionForgot
		if ch.met {
			action = ActionAnswered
		}
		ev := Event{
			Actor:  ActorCity,
			Action: action,
			Target: ch.door,
			Args:   map[string]any{"value": ch.value},
			Result: "ok",
		}
		res := finishResult(s.state, ev, ControlResult{
			OK: true, Actor: ev.Actor, Action: ev.Action, Target: ev.Target,
		}, locale)
		ev.Narration = res.Narration
		s.state.EventHistory = append(s.state.EventHistory, ev)
		const maxHistory = 20
		if len(s.state.EventHistory) > maxHistory {
			s.state.EventHistory = s.state.EventHistory[len(s.state.EventHistory)-maxHistory:]
		}
	}
	err := s.persistLocked()
	s.mu.Unlock()
	if err != nil {
		log.Printf("[mywant-sse] persist: %v", err)
	}
}

// seedDoorSatisfiedLocked records what each board-reading door's condition says
// right now, without narrating any of it.
//
// This is the baseline, and it has to be taken on arrival rather than on the
// first thing_changed to reach us. Taking it lazily meant the first change after
// entering a stage was spent establishing what "before" was — so the player
// pinned a value, which is the whole move of fortress2, and the game recorded it
// silently and said nothing.
//
// Called with s.mu held, from wherever the stage in play changes.
func (s *Server) seedDoorSatisfiedLocked() {
	s.doorSatisfied = map[string]bool{}
	stage := s.state.Stages[s.state.CurrentStage]
	if stage == nil {
		return
	}
	for id, d := range stage.Doors {
		c, pinned := d.RequiresThingNamed, false
		if c == nil {
			c, pinned = d.RequiresThingPinned, true
		}
		if c == nil {
			continue
		}
		v := c.Want(stage)
		if pinned {
			s.doorSatisfied[id] = currentBoard.Pinned(c.Subtype, v)
		} else {
			s.doorSatisfied[id] = currentBoard.Named(c.Subtype, v)
		}
	}
}
