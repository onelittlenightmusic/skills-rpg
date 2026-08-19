package server

import "fmt"

// A stage's spine, written once.
//
// Every stage in this game is the same shape: a line of rooms, a door between
// each pair, a reader powering each door, and the player walking from the first
// room to the last. Written out longhand that is a waypoint block with the
// adjacency spelled both ways, a door block repeating the two room ids, an
// achievement for every reader coming alive, an achievement for every room
// entered, and a cleared_when naming the last of them — thirty-odd lines of
// bookkeeping per stage, in which the only interesting words are the ids.
//
// So a stage may instead write its rooms in order:
//
//	rooms:
//	  - { id: cistern_floor, label: "The Cistern Floor" }
//	  - { id: cistern_stair, label: "The Stair Out",
//	      door: { id: cistern_hatch, device: cistern_gauge } }
//
// and this derives the rest. What it derives is exactly what was being typed:
// nothing here is a new rule, and a stage that wants a shape this cannot
// describe — a loop, a room off to one side, a door that starts open — writes
// `waypoints:` and `doors:` longhand as before. The two forms compose; anything
// written by hand wins.
//
// The generated achievement ids are part of the contract, because goals and
// narrations refer to them:
//
//	<device>_running  — that reader answered the city
//	reached_<room>    — the player got into that room
//
// and cleared_when defaults to reaching the last room, which is what finishing
// a stage has meant in every stage so far.

// Room is one room in a stage's spine, and the door you get into it by.
type Room struct {
	ID    string    `yaml:"id" json:"id"`
	Label string    `yaml:"label,omitempty" json:"label,omitempty"`
	Door  *RoomDoor `yaml:"door,omitempty" json:"door,omitempty"`
}

// RoomDoor is the door between this room and the one before it. Omit it and the
// two rooms simply join.
type RoomDoor struct {
	ID              string `yaml:"id" json:"id"`
	Device          string `yaml:"device,omitempty" json:"device,omitempty"`
	Key             string `yaml:"key,omitempty" json:"key,omitempty"`
	BlockedByDevice string `yaml:"blocked_by_device,omitempty" json:"blocked_by_device,omitempty"`
}

// DeviceRunning is the achievement awarded when a reader answers the city.
func DeviceRunning(deviceID string) string { return deviceID + "_running" }

// RoomReached is the achievement awarded when the player enters a room.
func RoomReached(roomID string) string { return "reached_" + roomID }

// ExpandRooms fills in the longhand a stage's `rooms:` stands for.
//
// Additive and never destructive: a waypoint, door or achievement the stage
// wrote itself is left exactly as written, because the shorthand exists to save
// typing and not to overrule the author.
func ExpandRooms(st *Stage) error {
	if len(st.Rooms) == 0 {
		return nil
	}
	if st.Waypoints == nil {
		st.Waypoints = map[string]*Waypoint{}
	}
	if st.Doors == nil {
		st.Doors = map[string]*Door{}
	}

	for i, r := range st.Rooms {
		if r.ID == "" {
			return fmt.Errorf("%s: room %d has no id", st.ID, i)
		}
		var adjacent []string
		if i > 0 {
			adjacent = append(adjacent, st.Rooms[i-1].ID)
		}
		if i+1 < len(st.Rooms) {
			adjacent = append(adjacent, st.Rooms[i+1].ID)
		}
		if _, written := st.Waypoints[r.ID]; !written {
			st.Waypoints[r.ID] = &Waypoint{Label: r.Label, Adjacent: adjacent}
		}
		if r.Door == nil {
			continue
		}
		if i == 0 {
			return fmt.Errorf("%s: the first room (%s) has a door, but nothing to be a door from", st.ID, r.ID)
		}
		if r.Door.ID == "" {
			return fmt.Errorf("%s: the door into %s has no id", st.ID, r.ID)
		}
		if _, written := st.Doors[r.Door.ID]; !written {
			st.Doors[r.Door.ID] = &Door{
				Between:         [2]string{st.Rooms[i-1].ID, r.ID},
				Open:            false,
				Locked:          true,
				Key:             r.Door.Key,
				RequiresDevice:  r.Door.Device,
				BlockedByDevice: r.Door.BlockedByDevice,
			}
		}
	}

	defined := map[string]bool{}
	for _, ad := range st.AchievementDefs {
		defined[ad.ID] = true
	}
	add := func(id string, when AchievementMatcher) {
		if id == "" || defined[id] {
			return
		}
		defined[id] = true
		st.AchievementDefs = append(st.AchievementDefs, AchievementDef{ID: id, When: when})
	}

	for i, r := range st.Rooms {
		if r.Door != nil && r.Door.Device != "" {
			add(DeviceRunning(r.Door.Device), AchievementMatcher{
				Actor: ActorCity, Action: ActionAnswered, Target: r.Door.Device,
			})
		}
		if i == 0 {
			continue
		}
		add(RoomReached(r.ID), AchievementMatcher{
			Actor: ActorYou, Action: ActionMove, Target: r.ID, Result: "ok",
		})
	}

	if st.InitialPosition == "" {
		st.InitialPosition = st.Rooms[0].ID
	}
	if st.ClearedWhen == "" && len(st.Rooms) > 1 {
		st.ClearedWhen = RoomReached(st.Rooms[len(st.Rooms)-1].ID)
	}
	return nil
}
