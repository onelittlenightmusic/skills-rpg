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
// It opens the doors it is watching, and that is the point rather than an
// overreach. In the dungeon chap opens doors because `you` has no privilege to;
// a fortress gate is different — it answers to its maker, and what was missing
// was the name, not somebody to push it. Requiring the player to name a value
// and then ask chap to go and try the door adds a step that teaches nothing and
// dilutes the one thing this world is about: naming is the act, and the gate
// opening is its consequence.
//
// It still moves nobody and clears nothing. Walking is the player's.

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
// nudge asks for a reconciliation without waiting for one. Non-blocking: a
// burst of events is one look at the board, not one per event.
func (s *Server) nudge() {
	select {
	case s.reconcile <- struct{}{}:
	default:
	}
}

// How often the board is looked at when nothing has said anything. A backstop,
// not the mechanism: mywant does not emit an event for every fact an activator
// can read — a relation drawn between two wants is the case that proved it — and
// a door that only ever opens when something happened to be announced is a door
// that sometimes does not open.
const boardSweepInterval = 5 * time.Second

// reconcileLoop is the only caller of reconcileBoardDoors. Serialising the
// passes here means a burst of events cannot have two of them reading the board
// at once, and the pause afterwards is what makes a burst cheap.
func (s *Server) reconcileLoop() {
	tick := time.NewTicker(boardSweepInterval)
	defer tick.Stop()
	for {
		select {
		case <-s.reconcile:
		case <-tick.C:
		}
		s.reconcileBoardDoors()
		time.Sleep(500 * time.Millisecond)
	}
}

func (s *Server) watchMywant() {
	if s.cfg.MywantURL == "" {
		return
	}
	go s.reconcileLoop()
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
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "data:"):
			// Neither the payload nor the event type is read. Which thing
			// changed does not matter — what matters is whether any activator's
			// answer is now different, and that is a question about the board as
			// a whole, asked below. Reading the id and trusting it would also be
			// wrong: a value can be named by something other than the event that
			// named it.
			//
			// And not `thing_changed` alone, which is what this used to wait
			// for. Act 1's conditions are about named values, so that was the
			// whole story while Act 1 was the whole world; Acts 2 and 3 read
			// want state and the roads between wants, and connecting two wants
			// produced no thing at all — the activator came alive and the door
			// in front of it sat shut, because nothing had told the game to look.
			s.nudge()
		}
	}
	return sc.Err()
}

// reconcileBoardDoors records an event for each registry-reading device whose
// answer has changed since the last look — running, or no longer running.
//
// Only to narrate it. The device answers from the city whenever it is asked
// (Device.IsOn), and the door in front of it reads the device exactly as every
// other door in this game reads one — so nothing here opens or switches
// anything. What a stream of changes is good for is knowing that something
// CHANGED, which a derived value cannot tell you on its own, and which is what
// a story needs.
//
// Both directions, because a name can come and go: pinned, unpinned, deleted,
// named again. And no stage anywhere in here — the city has no stages, and
// scoping this to the one the game thinks it is in was how a device on the same
// canvas, one room over, stayed silent when the value it watched was pinned.
func (s *Server) reconcileBoardDoors() {
	s.mu.Lock()
	type change struct {
		device, value string
		on            bool
	}
	var news []change
	for _, stage := range s.state.Stages {
		for id, dev := range stage.Devices {
			if dev.cond() == nil {
				continue
			}
			on := dev.IsOn()
			// Keyed by stage as well as by device: both fortress stages have a
			// device called registry_terminal, and one map entry for the two of
			// them meant the first stage's change ate the second's — the yard
			// door sat shut with its reader plainly running.
			key := stage.ID + "/" + id
			if was, seen := s.doorSatisfied[key]; seen && was == on {
				continue
			}
			if s.doorSatisfied == nil {
				s.doorSatisfied = map[string]bool{}
			}
			s.doorSatisfied[key] = on
			// The gate takes power, and swings. World 1 needs chap here because
			// `you` has no privilege to work a door; a fortress gate answers its
			// maker, and what was missing was the name — so once the reader is
			// running there is nobody left to wait for. Losing the name shuts it
			// again, which is the same sentence read backwards.
			for _, d := range stage.Doors {
				if d.RequiresDevice == id {
					d.Open = on
					d.Locked = !on
				}
			}
			news = append(news, change{device: id, value: dev.Subject(), on: on})
		}
	}
	if len(news) == 0 {
		s.mu.Unlock()
		return
	}
	locale := s.currentLocaleLocked()
	for _, ch := range news {
		action := ActionForgot
		if ch.on {
			action = ActionAnswered
		}
		ev := Event{
			Actor:  ActorCity,
			Action: action,
			Target: ch.device,
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

// seedDoorSatisfiedLocked records what every registry reader says right now, so
// the next change can be told from a repeat. Nothing applied, nothing narrated.
// Called with s.mu held.
func (s *Server) seedDoorSatisfiedLocked() {
	s.doorSatisfied = map[string]bool{}
	for _, stage := range s.state.Stages {
		for id, dev := range stage.Devices {
			if dev.cond() == nil {
				continue
			}
			on := dev.IsOn()
			s.doorSatisfied[stage.ID+"/"+id] = on
			for _, d := range stage.Doors {
				if d.RequiresDevice == id {
					d.Open = on
					d.Locked = !on
				}
			}
		}
	}
}
