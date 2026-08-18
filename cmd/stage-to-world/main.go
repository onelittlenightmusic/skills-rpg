// stage-to-world converts skills-rpg's stage1..stage9 waypoint/door graphs
// into a single mywant "world" (a canvas layout built from `wall`/`rpg_door`/
// device wants) named "skills-rpg" by default.
//
// Every stage's waypoints are flattened into a single left-to-right room
// chain (all 9 stages are simple linear graphs — verified against the real
// stage1..stage9 data, no branching), one row per stage so stages never
// overlap. Each room is a 5x5 tile footprint (walls included); adjacent
// rooms within a stage share their boundary wall. A skills-rpg `Door` between
// two rooms becomes a 1x1 `rpg_door` want in a gap in that shared wall, and
// every entry in a stage's `devices` map becomes an `rpg_alarm` or
// `rpg_generator` tile on the floor in front of the door it gates — both
// mirror the running rpg-server (stage_id/device_id params — see
// server/skills/rpg-door and server/skills/rpg-device). A plain `Adjacent`
// link with no `Door` entry is left fully open (no wall on that side at all).
//
// Usage:
//
//	go run ./cmd/stage-to-world [-stages-dir server/stages] [-world skills-rpg] [-mywant-url http://localhost:8080]
//
// -dry-run prints the wants it would import and touches nothing, which is the
// way to see a layout change before it replaces a world someone is playing in.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/onelittlenightmusic/skills-rpg/server"
	"gopkg.in/yaml.v3"
)

// Room geometry: a roomSize x roomSize (6x6) box per room. Left/right walls
// run the FULL height (anchored at the top corner, length=sideLen) so a
// shared boundary column is exactly one want, reused by both neighbors.
// Top/bottom walls are shortened to exclude both corner cells (they're
// already claimed by the left/right walls) so no two wall wants ever share a
// cell — see the WantCanvas.tsx positionMap fix this pairs with, but this
// also just produces a cleaner, unambiguous room outline on its own.
const (
	roomSize  = 6            // full room footprint (walls included), tiles
	roomPitch = roomSize - 1 // distance between room anchors; rooms share one wall column
	sideLen   = roomSize - 1 // tileFootprintCells span = length+1: full side (6 cells) needs length=5
	topBotLen = roomSize - 3 // top/bottom span (4 cells, corners excluded) needs length=3
	rowGap    = 10           // vertical spacing between stage rows (> roomSize, so rows never touch)
)

// want mirrors just enough of mywant's want-spec YAML shape to be accepted
// by POST /api/v1/wants/import — no dependency on mywant's own Go module.
type want struct {
	Metadata wantMetadata `yaml:"metadata"`
	Spec     wantSpec     `yaml:"spec"`
}

type wantMetadata struct {
	ID     string            `yaml:"id"`
	Name   string            `yaml:"name"`
	Type   string            `yaml:"type"`
	Labels map[string]string `yaml:"labels,omitempty"`
}

type wantSpec struct {
	Params map[string]any `yaml:"params"`
}

func main() {
	stagesDir := flag.String("stages-dir", "server/stages", "directory containing stageN.yaml files")
	worldName := flag.String("world", "skills-rpg", "target mywant world name")
	mywantURL := flag.String("mywant-url", envOr("MYWANT_URL", "http://localhost:8080"), "mywant server base URL")
	onlyAddMissing := flag.Bool("only-add-missing", false, "reconcile mode: add generated wants whose ID doesn't already exist in the currently active world, without opening a different world or clearing anything (safe to run against a world that has extra manually-deployed wants, e.g. a live coding-instance)")
	rpgWorld := flag.String("rpg-world", "", "skills-rpg world this board is for; the generated board carries an rpg_world want that keeps the game in it (default: the stages dir's basename, or none for the flat dungeon dir)")
	dryRun := flag.Bool("dry-run", false, "print the generated wants as YAML and exit, without opening, clearing or writing to any world")
	flag.Parse()

	stages, err := loadStages(*stagesDir)
	if err != nil {
		fatalf("load stages: %v", err)
	}
	if len(stages) == 0 {
		fatalf("no stageN.yaml files found in %s", *stagesDir)
	}

	var wants []want
	layouts := make(map[string]stageInfo, len(stages)) // stage.ID -> its room layout info
	for i, stage := range stages {
		roomWants, info := layoutStage(stage, i)
		wants = append(wants, roomWants...)
		layouts[stage.ID] = info
	}

	// Second pass: a `startpoint` in every stage's first room, and a
	// `next_stage` portal in every stage's last room pointing at the next
	// stage's startpoint (skipped where NextStage is empty or unknown, e.g.
	// the final stage). Needs all stages laid out first since a stage's
	// portal targets the *next* stage's coordinates.
	// Who leads into each stage, read back off the same NextStage links the
	// portals are built from — so the way back cannot disagree with the way
	// forward. A stage nobody points at (the first one) simply has no entry
	// here, and its start point is left with no way back, which is the truth.
	prevOf := map[string]string{}
	for _, stage := range stages {
		if stage.NextStage != "" {
			prevOf[stage.NextStage] = stage.ID
		}
	}

	for _, stage := range stages {
		info := layouts[stage.ID]
		startParams := map[string]any{"label": fmt.Sprintf("%s entrance", stage.ID)}
		// Landing on the portal you came through, not on that stage's own start
		// point: going back should put you where you left, facing the door you
		// walked into. (Its start point is where you would land arriving from
		// the stage before it — a different journey.)
		if prev, ok := prevOf[stage.ID]; ok {
			if prevInfo, ok := layouts[prev]; ok {
				startParams["rpg_prev_stage_id"] = prev
				startParams["target_x"] = prevInfo.lastX
				startParams["target_y"] = prevInfo.lastY
			}
		}
		wants = append(wants, want{
			Metadata: wantMetadata{
				ID:   fmt.Sprintf("want-%s-startpoint", stage.ID),
				Name: fmt.Sprintf("%s-startpoint", stage.ID),
				Type: "startpoint",
				Labels: map[string]string{
					"mywant.io/canvas-x": fmt.Sprint(info.startX),
					"mywant.io/canvas-y": fmt.Sprint(info.startY),
				},
			},
			Spec: wantSpec{Params: startParams},
		})

		if stage.NextStage == "" {
			continue
		}
		nextInfo, ok := layouts[stage.NextStage]
		if !ok {
			continue // NextStage points outside the converted set — nothing to target
		}
		wants = append(wants, want{
			Metadata: wantMetadata{
				ID:   fmt.Sprintf("want-%s-next-stage", stage.ID),
				Name: fmt.Sprintf("%s-next-stage", stage.ID),
				Type: "next_stage",
				Labels: map[string]string{
					"mywant.io/canvas-x": fmt.Sprint(info.lastX),
					"mywant.io/canvas-y": fmt.Sprint(info.lastY),
				},
			},
			Spec: wantSpec{Params: map[string]any{
				"rpg_next_stage_id": stage.NextStage,
				"target_x":          nextInfo.startX,
				"target_y":          nextInfo.startY,
			}},
		})
	}
	// One want that keeps the game in the world this board is for. It sits
	// with the board rather than being something the player has to run,
	// because "which world skills-rpg is in" is a property of which board is
	// open — see server/skills/rpg-world.
	if w := resolveRPGWorld(*rpgWorld, *stagesDir); w != "" {
		wants = append(wants, want{
			Metadata: wantMetadata{
				ID:   fmt.Sprintf("want-%s-world", w),
				Name: fmt.Sprintf("%s-world", w),
				Type: "rpg_world",
				Labels: map[string]string{
					// Off the walked area: it is machinery, not scenery.
					"mywant.io/canvas-x": "-3",
					"mywant.io/canvas-y": "-3",
				},
			},
			Spec: wantSpec{Params: map[string]any{"world": w}},
		})
	}

	fmt.Printf("Generated %d wants (wall+door+device+startpoint+next_stage+rpg_world) from %d stages\n", len(wants), len(stages))

	if *dryRun {
		data, err := yaml.Marshal(wants)
		if err != nil {
			fatalf("marshal wants: %v", err)
		}
		os.Stdout.Write(data)
		return
	}

	if *onlyAddMissing {
		existing, err := existingWantIDs(*mywantURL)
		if err != nil {
			fatalf("list existing wants: %v", err)
		}
		var missing []want
		for _, w := range wants {
			if !existing[w.Metadata.ID] {
				missing = append(missing, w)
			}
		}
		fmt.Printf("%d of those already exist; adding %d missing want(s)\n", len(wants)-len(missing), len(missing))
		if len(missing) == 0 {
			return
		}
		if err := importWants(*mywantURL, missing); err != nil {
			fatalf("import missing wants: %v", err)
		}
		if err := waitForWantCount(*mywantURL, len(existing)+len(missing)); err != nil {
			fatalf("waiting for import to complete: %v", err)
		}
		if err := saveWorld(*mywantURL, *worldName); err != nil {
			fatalf("save world %q: %v", *worldName, err)
		}
		fmt.Printf("Saved world %q\n", *worldName)
		return
	}

	if err := openWorld(*mywantURL, *worldName); err != nil {
		fatalf("open world %q: %v", *worldName, err)
	}
	// Re-running this tool against a world it already populated would
	// otherwise collide on the (deterministic) want names openWorld just
	// reloaded from <world>.yaml — clear it back to empty first so every run
	// is idempotent.
	if err := clearAllWants(*mywantURL); err != nil {
		fatalf("clear existing world contents: %v", err)
	}
	if err := importWants(*mywantURL, wants); err != nil {
		fatalf("import wants: %v", err)
	}
	// POST /api/v1/wants/import responds before the server finishes
	// registering every want (it finalizes them in a background goroutine) —
	// wait for the live count to catch up before snapshotting, or `save`
	// below can persist an empty/partial world.
	if err := waitForWantCount(*mywantURL, len(wants)); err != nil {
		fatalf("waiting for import to complete: %v", err)
	}
	if err := saveWorld(*mywantURL, *worldName); err != nil {
		fatalf("save world %q: %v", *worldName, err)
	}
	fmt.Printf("Saved world %q with %d wants\n", *worldName, len(wants))
}

// loadStages reads stage1.yaml..stage9.yaml (in filename order, which sorts
// numerically for single-digit stage counts) directly into server.Stage —
// the same exported type rpg-server itself uses, so this stays in sync with
// the real schema. Extra top-level YAML keys (e.g. "locales") are silently
// ignored by yaml.Unmarshal since server.Stage has no strict-mode decoding.
func loadStages(dir string) ([]*server.Stage, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".yaml") {
			continue
		}
		if strings.Contains(name, "-jp") {
			continue // locale-only file
		}
		// world.yaml describes the world a directory of stages belongs to. It
		// sits beside them and is not one.
		if name == "world.yaml" {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)

	byID := map[string]*server.Stage{}
	var order []string
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return nil, err
		}
		var st server.Stage
		if err := yaml.Unmarshal(data, &st); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		if st.ID == "" || len(st.Waypoints) == 0 {
			continue
		}
		byID[st.ID] = &st
		order = append(order, st.ID)
	}
	if len(byID) == 0 {
		return nil, nil
	}

	// A world is the chain its stages form, not every stage file in the
	// directory. server/stages also holds workshop1.yaml, which has an id and a
	// waypoint and is nobody's next_stage — it is a workshop, not part of the
	// dungeon, and putting its room on the dungeon's board would be wrong.
	//
	// This used to be handled by requiring the filename to start with "stage",
	// which worked while there was one world whose stages were numbered. A
	// second world names its stages after itself, so the rule had to become one
	// about the stages rather than about their filenames — and walking the
	// chain is what "which stages are this world's" actually means.
	first := firstStage(dir, order)
	var chain []*server.Stage
	seen := map[string]bool{}
	for id := first; id != "" && !seen[id]; {
		st, ok := byID[id]
		if !ok {
			break // names a stage that is not written yet
		}
		seen[id] = true
		chain = append(chain, st)
		id = st.NextStage
	}
	return chain, nil
}

// firstStage is where the chain starts: world.yaml says so when the directory
// has one, and otherwise it is the earliest id in sorted order — which is what
// the dungeon has always meant by stage1.
func firstStage(dir string, order []string) string {
	if data, err := os.ReadFile(filepath.Join(dir, "world.yaml")); err == nil {
		var w struct {
			FirstStage string `yaml:"first_stage"`
		}
		if yaml.Unmarshal(data, &w) == nil && w.FirstStage != "" {
			return w.FirstStage
		}
	}
	sorted := append([]string(nil), order...)
	sort.Strings(sorted)
	if len(sorted) == 0 {
		return ""
	}
	return sorted[0]
}

// walkWaypoints flattens a stage's waypoint graph into a single visiting
// order, starting at InitialPosition and always stepping to the first
// unvisited Adjacent neighbor. All of stage1..stage9's real data is a simple
// chain (verified: max 2 Adjacent per waypoint, and that only at a single
// pass-through node), so this always reconstructs the full linear sequence.
func walkWaypoints(stage *server.Stage) []string {
	visited := map[string]bool{}
	var order []string
	cur := stage.InitialPosition
	for cur != "" && !visited[cur] {
		visited[cur] = true
		order = append(order, cur)
		next := ""
		if wp := stage.Waypoints[cur]; wp != nil {
			for _, adj := range wp.Adjacent {
				if !visited[adj] {
					next = adj
					break
				}
			}
		}
		cur = next
	}
	return order
}

// findDoor looks up the Door (if any) directly connecting waypoints a and b,
// regardless of which order Between lists them in.
func findDoor(stage *server.Stage, a, b string) (string, *server.Door) {
	for id, d := range stage.Doors {
		if (d.Between[0] == a && d.Between[1] == b) || (d.Between[0] == b && d.Between[1] == a) {
			return id, d
		}
	}
	return "", nil
}

// cell is a canvas square, as (x, y). Used as a map key, so devices can be
// laid down without ever landing two on the same square.
type cell struct{ x, y int }

// stageInfo records where a stage's first and last rooms' interiors are, so
// a later pass can place a startpoint in the first and a next_stage portal
// (targeting the next stage's startpoint) in the last. It also records the
// square every door tile ended up in, which is what device placement reads:
// a device stands in front of the door it gates, so it has to know where the
// door is.
type stageInfo struct {
	startX, startY int             // first room's interior center — where CursorMan should land
	lastX, lastY   int             // last room's interior center — where this stage's next_stage portal sits
	baseY          int             // top wall row of every room in this stage
	rightX         int             // right wall column of the last room — the stage's east edge
	doors          map[string]cell // rpg door id -> the square its tile stands in
}

// isFloor reports whether (x, y) is an interior square of one of this stage's
// rooms — inside the top/bottom walls, and not on one of the shared wall
// columns the rooms are separated by (every multiple of roomPitch).
func (s stageInfo) isFloor(c cell) bool {
	if c.y <= s.baseY || c.y >= s.baseY+roomSize-1 {
		return false
	}
	if c.x <= 0 || c.x >= s.rightX {
		return false
	}
	return c.x%roomPitch != 0
}

// layoutStage places stage's rooms in a single row (y = stageIndex*rowGap),
// left to right in waypoint-visit order.
func layoutStage(stage *server.Stage, stageIndex int) ([]want, stageInfo) {
	order := walkWaypoints(stage)
	baseY := stageIndex * rowGap

	var out []want
	doors := map[string]cell{}
	for i, wpID := range order {
		rx := i * roomPitch
		ry := baseY
		roomWants, doorKey, doorAt := roomWalls(stage, wpID, order, i, rx, ry)
		out = append(out, roomWants...)
		if doorKey != "" {
			doors[doorKey] = doorAt
		}
	}

	lastRX := (len(order) - 1) * roomPitch
	info := stageInfo{
		startX: 2, startY: baseY + 2, // first room's anchor is always rx=0
		lastX: lastRX + 2, lastY: baseY + 2,
		baseY: baseY, rightX: lastRX + roomSize - 1,
		doors: doors,
	}
	return append(out, deviceWants(stage, info)...), info
}

// roomWalls emits the wall(s) (and, where a Door connects to the next room,
// the door) bounding room i (anchored at rx,ry, a roomSize x roomSize box).
// It also returns that door's id and the square it stands in (empty id when
// the room has no door on its right side), so devices can be placed against
// it later.
func roomWalls(stage *server.Stage, wpID string, order []string, i, rx, ry int) ([]want, string, cell) {
	idPrefix := fmt.Sprintf("%s-%s", stage.ID, wpID)
	var out []want

	addWall := func(suffix string, x, y, rotation, length int) {
		out = append(out, want{
			Metadata: wantMetadata{
				ID:   fmt.Sprintf("want-%s-%s", idPrefix, suffix),
				Name: fmt.Sprintf("%s-%s", idPrefix, suffix),
				Type: "wall",
				Labels: map[string]string{
					"mywant.io/canvas-x":        fmt.Sprint(x),
					"mywant.io/canvas-y":        fmt.Sprint(y),
					"mywant.io/canvas-rotation": fmt.Sprint(rotation),
					"mywant.io/canvas-length":   fmt.Sprint(length),
				},
			},
			Spec: wantSpec{Params: map[string]any{}},
		})
	}

	// Top and bottom sides span only the middle 4 cells (x=rx+1..rx+4) — the
	// two corner cells on each row belong to the left/right walls instead,
	// so no two wall wants ever claim the same cell.
	addWall("top", rx+1, ry, 0, topBotLen)
	addWall("bottom", rx+1, ry+roomSize-1, 0, topBotLen)

	// Left side: only the first room in the stage gets one — every later
	// room's left side is the previous room's right side (shared wall).
	if i == 0 {
		addWall("left", rx, ry, 90, sideLen)
	}

	rightX := rx + roomSize - 1
	if i == len(order)-1 {
		// Last room in the stage: cap it with a full right wall.
		addWall("right", rightX, ry, 90, sideLen)
		return out, "", cell{}
	}

	nextID := order[i+1]
	doorKey, door := findDoor(stage, wpID, nextID)
	if door == nil {
		// Plain Adjacent with no Door entry — leave the middle of this shared
		// column fully open (free passage), but still cap both corners with a
		// 1x1 wall. Top/bottom walls only cover the middle 4 cells of a room
		// (corners are normally claimed by the left/right wall instead), so
		// skipping the vertical wall here entirely would leave a 1-cell hole
		// in the top and bottom wall lines at this column, not just an open
		// side passage.
		addWall("right-corner-top", rightX, ry, 0, 0)
		addWall("right-corner-bottom", rightX, ry+roomSize-1, 0, 0)
		return out, "", cell{}
	}

	// Split the shared (6-cell) wall into "2 cells + 1-cell door gap + 3
	// cells" and place a 1x1 door want in the gap, synced to this exact
	// rpg-server door. None of these three pieces share a cell with each
	// other or with any other wall in the room.
	//
	// The tile is `rpg_door`, not the bundled `door`: both mirror `locked`
	// from rpg-server and both block CursorMan while it is set, but rpg_door
	// also carries the reason the door is shut — which key opens it, whether
	// chap holds that key, which device is holding it — and that is the whole
	// point of the device tiles placed next to it.
	addWall("right-a", rightX, ry, 90, 1)
	addWall("right-b", rightX, ry+3, 90, 2)
	at := cell{rightX, ry + 2}
	out = append(out, want{
		Metadata: wantMetadata{
			ID:   fmt.Sprintf("want-%s-door-%s", idPrefix, doorKey),
			Name: fmt.Sprintf("%s-door-%s", idPrefix, doorKey),
			Type: "rpg_door",
			Labels: map[string]string{
				"mywant.io/canvas-x": fmt.Sprint(at.x),
				"mywant.io/canvas-y": fmt.Sprint(at.y),
			},
		},
		Spec: wantSpec{Params: map[string]any{
			"locked":   door.Locked,
			"stage_id": stage.ID,
			"door_id":  doorKey,
		}},
	})
	return out, doorKey, at
}

// Where a device tile is looked for, relative to the square it wants: the
// square itself, then straight back from the door, then forward, then a
// column further from the door. A room only ever holds a couple of devices,
// so this never has to search far — it exists so two devices gating the same
// door, or a device whose preferred square is already the startpoint, still
// land on floor rather than on top of something.
var deviceOffsets = []cell{
	{0, 0}, {0, -1}, {0, 1}, {0, -2}, {0, 2},
	{-1, 0}, {-1, -1}, {-1, 1}, {-2, 0}, {-2, -1}, {-2, 1},
}

// deviceWants places one tile per entry in the stage's `devices` map.
//
// Nothing in the stage data says where a device is — a device is only an id,
// a label and an on/off flag. What the data does say is what each device
// gates: a door names it in requires_device or blocked_by_device, and a
// device names whatever holds it shut in blocked_by_device. That is enough to
// place them. The device gating a door stands on the floor square in front of
// that door, and whatever holds that device shut stands one square further
// back, so the room reads in the order it has to be solved: the far tile
// first, then the near one, then the door.
func deviceWants(stage *server.Stage, info stageInfo) []want {
	if len(stage.Devices) == 0 {
		return nil
	}
	devIDs := slices.Sorted(maps.Keys(stage.Devices))
	controllable := devicesControllable(stage)

	// Which door each device gates. Doors are walked in id order so a device
	// gating more than one door still lands somewhere deterministic.
	gated := map[string]string{}
	for _, doorKey := range slices.Sorted(maps.Keys(stage.Doors)) {
		d := stage.Doors[doorKey]
		for _, devID := range []string{d.RequiresDevice, d.BlockedByDevice} {
			if devID == "" {
				continue
			}
			if _, taken := gated[devID]; !taken {
				gated[devID] = doorKey
			}
		}
	}

	// The startpoint and the next_stage portal are placed by the caller's
	// second pass, and they are floor squares like any other.
	taken := map[cell]bool{
		{info.startX, info.startY}: true,
		{info.lastX, info.lastY}:   true,
	}
	placed := map[string]cell{}
	var out []want

	put := func(devID string, want cell) bool {
		for _, off := range deviceOffsets {
			c := cell{want.x + off.x, want.y + off.y}
			if !info.isFloor(c) || taken[c] {
				continue
			}
			taken[c] = true
			placed[devID] = c
			out = append(out, deviceWant(stage, devID, c, controllable))
			return true
		}
		return false
	}

	// In front of the door each device gates.
	for _, devID := range devIDs {
		door, ok := info.doors[gated[devID]]
		if !ok {
			continue // no door, or a doorway the layout drew as an open gap
		}
		put(devID, cell{door.x - 1, door.y})
	}

	// Behind whatever they hold shut. Looped until nothing new lands, so a
	// chain of blockers keeps stepping one square further back.
	for progress := true; progress; {
		progress = false
		for _, devID := range devIDs {
			if _, done := placed[devID]; done {
				continue
			}
			for _, other := range devIDs {
				at, ok := placed[other]
				if !ok || stage.Devices[other].BlockedByDevice != devID {
					continue
				}
				progress = put(devID, cell{at.x, at.y - 1})
				break
			}
		}
	}

	// Anything left gates nothing the layout drew — park it in the first
	// room, off the line the player walks from the startpoint to the door.
	for _, devID := range devIDs {
		if _, done := placed[devID]; !done {
			put(devID, cell{info.startX, info.startY - 1})
		}
	}
	return out
}

func deviceWant(stage *server.Stage, devID string, at cell, controllable bool) want {
	slug := strings.ReplaceAll(devID, "_", "-")
	return want{
		Metadata: wantMetadata{
			ID:   fmt.Sprintf("want-%s-%s", stage.ID, slug),
			Name: fmt.Sprintf("%s-%s", stage.ID, slug),
			Type: deviceWantType(stage, devID),
			Labels: map[string]string{
				"mywant.io/canvas-x": fmt.Sprint(at.x),
				"mywant.io/canvas-y": fmt.Sprint(at.y),
			},
		},
		Spec: wantSpec{Params: map[string]any{
			"stage_id":     stage.ID,
			"device_id":    devID,
			"controllable": controllable,
		}},
	}
}

// deviceWantType picks the tile a device is drawn as. The stage data has no
// type field, but it does say which way the device has to be pointed for the
// stage to open up, and that is the difference between the two tiles: an
// alarm is a thing you switch off (something else is held shut while it
// runs), a generator is a thing you switch on (something else needs it
// running). Where a device gates nothing, its resting state says the same —
// alarms start on, generators start off.
func deviceWantType(stage *server.Stage, devID string) string {
	for _, d := range stage.Doors {
		if d.BlockedByDevice == devID {
			return "rpg_alarm"
		}
	}
	for _, other := range stage.Devices {
		if other.BlockedByDevice == devID {
			return "rpg_alarm"
		}
	}
	for _, d := range stage.Doors {
		if d.RequiresDevice == devID {
			return "rpg_generator"
		}
	}
	if stage.Devices[devID].On {
		return "rpg_alarm"
	}
	return "rpg_generator"
}

// devicesControllable reports whether the stage's device tiles may be clicked
// to have chap operate them.
//
// Usually yes — operating the device from the board is the point. The
// exception is a stage whose own goals tell the player to deploy a want to do
// it (stage9): there a clickable tile is the answer sheet, so those tiles are
// left as read-only mirrors of the server.
func devicesControllable(stage *server.Stage) bool {
	for _, rule := range stage.NextGoalRules {
		if strings.Contains(rule.Goal.RequiredSkill, "mywant-deploy") {
			return false
		}
	}
	return true
}

func openWorld(base, name string) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/worlds/%s/open", base, url.PathEscape(name)), nil)
	if err != nil {
		return err
	}
	return doRequest(req)
}

func saveWorld(base, name string) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/worlds/%s/save", base, url.PathEscape(name)), nil)
	if err != nil {
		return err
	}
	return doRequest(req)
}

func importWants(base string, wants []want) error {
	data, err := yaml.Marshal(wants)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, base+"/api/v1/wants/import", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/yaml")
	return doRequest(req)
}

// existingWantIDs returns the set of want IDs already present in the
// currently active world (used by -only-add-missing to avoid re-creating or
// colliding with anything, including manually-deployed wants this tool never
// generated in the first place). System wants are left out, so its size is
// the same population waitForWantCount counts; no generated id can name one
// anyway.
func existingWantIDs(base string) (map[string]bool, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	live, err := deletableWantIDs(client, base)
	if err != nil {
		return nil, err
	}
	ids := make(map[string]bool, len(live))
	for _, id := range live {
		ids[id] = true
	}
	return ids, nil
}

// clearAllWants deletes every deletable want in the active world (via the
// generic batch DELETE /api/v1/wants) and waits for the deletions to land. A
// no-op if the world is already empty.
//
// System wants (isSystemWant — `robot` is one) are left out of the batch.
// They have to be: the endpoint rejects the whole request with 403 if a
// single id in it names one, so including them deletes nothing at all. There
// is no reason to want them gone either — the server puts them back by
// itself. So the world never empties, and what this waits for is the
// deletable ones being gone rather than a count of zero: a delete landing
// mid-import would take a freshly imported want with it.
func clearAllWants(base string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	live, err := deletableWantIDs(client, base)
	if err != nil {
		return err
	}
	if len(live) == 0 {
		return nil
	}

	payload, err := json.Marshal(map[string]any{"ids": live})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, base+"/api/v1/wants", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if err := doRequest(req); err != nil {
		return err
	}

	for i := 0; i < 50; i++ {
		time.Sleep(200 * time.Millisecond)
		left, err := deletableWantIDs(client, base)
		if err != nil {
			continue
		}
		if len(left) == 0 {
			return nil
		}
	}
	return fmt.Errorf("timed out waiting for existing wants to clear")
}

// deletableWantIDs lists the ids of every want in the active world that the
// server will actually let go of.
func deletableWantIDs(client *http.Client, base string) ([]string, error) {
	resp, err := client.Get(base + "/api/v1/wants")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var body struct {
		Wants []struct {
			Metadata struct {
				ID       string `json:"id"`
				IsSystem bool   `json:"isSystemWant"`
			} `json:"metadata"`
		} `json:"wants"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	var ids []string
	for _, w := range body.Wants {
		if !w.Metadata.IsSystem {
			ids = append(ids, w.Metadata.ID)
		}
	}
	return ids, nil
}

// waitForWantCount polls until the world holds at least `want` deletable
// wants (bounded, ~10s), so a subsequent save doesn't race the import
// endpoint's background finalization. System wants are not counted — they are
// none of this tool's doing, and counting them would let the wait finish one
// want early.
func waitForWantCount(base string, want int) error {
	client := &http.Client{Timeout: 10 * time.Second}
	var lastCount int
	for i := 0; i < 50; i++ {
		ids, err := deletableWantIDs(client, base)
		if err == nil {
			lastCount = len(ids)
			if lastCount >= want {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %d wants (last saw %d)", want, lastCount)
}

func doRequest(req *http.Request) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s -> %d: %s", req.Method, req.URL, resp.StatusCode, string(body))
	}
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// resolveRPGWorld names the skills-rpg world a generated board is for.
//
// Explicit flag first; otherwise the stages directory's own name, which is how
// the server names worlds too. The flat stages/ directory is the dungeon —
// the inverse of the server's worldDir(), which maps the dungeon back to ".".
//
// The dungeon needs this want as much as any other board does. It is where the
// game starts, but a player standing in the fortress who opens the dungeon
// board expects to be in the dungeon, and only a want on that board can do it.
func resolveRPGWorld(flagVal, stagesDir string) string {
	if flagVal != "" {
		return flagVal
	}
	base := filepath.Base(filepath.Clean(stagesDir))
	if base == "stages" || base == "." || base == "/" {
		return "dungeon"
	}
	return base
}
