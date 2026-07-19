// stage-to-world converts skills-rpg's stage1..stage9 waypoint/door graphs
// into a single mywant "world" (a canvas layout built from `wall`/`door`
// wants) named "skills-rpg" by default.
//
// Every stage's waypoints are flattened into a single left-to-right room
// chain (all 9 stages are simple linear graphs — verified against the real
// stage1..stage9 data, no branching), one row per stage so stages never
// overlap. Each room is a 5x5 tile footprint (walls included); adjacent
// rooms within a stage share their boundary wall. A skills-rpg `Door` between
// two rooms becomes a 1x1 `door` want in a gap in that shared wall, wired to
// live-sync its "locked" state from the running rpg-server (rpg_stage_id/
// rpg_door_id params — see mywant's door.yaml). A plain `Adjacent` link with
// no `Door` entry is left fully open (no wall on that side at all).
//
// Usage:
//
//	go run ./cmd/stage-to-world [-stages-dir server/stages] [-world skills-rpg] [-mywant-url http://localhost:8080]
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	roomSize    = 6          // full room footprint (walls included), tiles
	roomPitch   = roomSize - 1 // distance between room anchors; rooms share one wall column
	sideLen     = roomSize - 1 // tileFootprintCells span = length+1: full side (6 cells) needs length=5
	topBotLen   = roomSize - 3 // top/bottom span (4 cells, corners excluded) needs length=3
	rowGap      = 10         // vertical spacing between stage rows (> roomSize, so rows never touch)
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
	for _, stage := range stages {
		info := layouts[stage.ID]
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
			Spec: wantSpec{Params: map[string]any{"label": fmt.Sprintf("%s entrance", stage.ID)}},
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
	fmt.Printf("Generated %d wants (wall+door+startpoint+next_stage) from %d stages\n", len(wants), len(stages))

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
		if !strings.HasPrefix(name, "stage") || strings.Contains(name, "-jp") {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)

	stages := make([]*server.Stage, 0, len(files))
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
			continue // not a real stage file (e.g. workshop1.yaml has a different shape)
		}
		stages = append(stages, &st)
	}
	return stages, nil
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

// stageInfo records where a stage's first and last rooms' interiors are, so
// a later pass can place a startpoint in the first and a next_stage portal
// (targeting the next stage's startpoint) in the last.
type stageInfo struct {
	startX, startY int // first room's interior center — where CursorMan should land
	lastX, lastY   int // last room's interior center — where this stage's next_stage portal sits
}

// layoutStage places stage's rooms in a single row (y = stageIndex*rowGap),
// left to right in waypoint-visit order.
func layoutStage(stage *server.Stage, stageIndex int) ([]want, stageInfo) {
	order := walkWaypoints(stage)
	baseY := stageIndex * rowGap

	var out []want
	for i, wpID := range order {
		rx := i * roomPitch
		ry := baseY
		out = append(out, roomWalls(stage, wpID, order, i, rx, ry)...)
	}

	lastRX := (len(order) - 1) * roomPitch
	return out, stageInfo{
		startX: 2, startY: baseY + 2, // first room's anchor is always rx=0
		lastX: lastRX + 2, lastY: baseY + 2,
	}
}

// roomWalls emits the wall(s) (and, where a Door connects to the next room,
// the door) bounding room i (anchored at rx,ry, a roomSize x roomSize box).
func roomWalls(stage *server.Stage, wpID string, order []string, i, rx, ry int) []want {
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
		return out
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
		return out
	}

	// Split the shared (6-cell) wall into "2 cells + 1-cell door gap + 3
	// cells" and place a 1x1 door want in the gap, synced to this exact
	// rpg-server door. None of these three pieces share a cell with each
	// other or with any other wall in the room.
	addWall("right-a", rightX, ry, 90, 1)
	addWall("right-b", rightX, ry+3, 90, 2)
	out = append(out, want{
		Metadata: wantMetadata{
			ID:   fmt.Sprintf("want-%s-door-%s", idPrefix, doorKey),
			Name: fmt.Sprintf("%s-door-%s", idPrefix, doorKey),
			Type: "door",
			Labels: map[string]string{
				"mywant.io/canvas-x": fmt.Sprint(rightX),
				"mywant.io/canvas-y": fmt.Sprint(ry + 2),
			},
		},
		Spec: wantSpec{Params: map[string]any{
			"locked":       door.Locked,
			"rpg_stage_id": stage.ID,
			"rpg_door_id":  doorKey,
		}},
	})
	return out
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
// generated in the first place).
func existingWantIDs(base string) (map[string]bool, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(base + "/api/v1/wants")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var body struct {
		Wants []struct {
			Metadata struct {
				ID string `json:"id"`
			} `json:"metadata"`
		} `json:"wants"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	ids := make(map[string]bool, len(body.Wants))
	for _, w := range body.Wants {
		ids[w.Metadata.ID] = true
	}
	return ids, nil
}

// clearAllWants deletes every want currently in the active world (via the
// generic batch DELETE /api/v1/wants, which only ever targets non-system
// wants — see mywant's listWants/exportableWants convention), then waits for
// the count to drop to zero. A no-op if the world is already empty.
func clearAllWants(base string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(base + "/api/v1/wants")
	if err != nil {
		return err
	}
	var body struct {
		Wants []struct {
			Metadata struct {
				ID string `json:"id"`
			} `json:"metadata"`
		} `json:"wants"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if len(body.Wants) == 0 {
		return nil
	}

	ids := make([]string, 0, len(body.Wants))
	for _, w := range body.Wants {
		ids = append(ids, w.Metadata.ID)
	}
	payload, err := json.Marshal(map[string]any{"ids": ids})
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
		resp, err := client.Get(base + "/api/v1/wants")
		if err == nil {
			var body struct {
				Wants []json.RawMessage `json:"wants"`
			}
			if json.NewDecoder(resp.Body).Decode(&body) == nil && len(body.Wants) == 0 {
				resp.Body.Close()
				return nil
			}
			resp.Body.Close()
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for existing wants to clear")
}

// waitForWantCount polls GET /api/v1/wants until it reports at least `want`
// wants (bounded, ~10s), so a subsequent save doesn't race the import
// endpoint's background finalization.
func waitForWantCount(base string, want int) error {
	client := &http.Client{Timeout: 10 * time.Second}
	var lastCount int
	for i := 0; i < 50; i++ {
		resp, err := client.Get(base + "/api/v1/wants")
		if err == nil {
			var body struct {
				Wants []json.RawMessage `json:"wants"`
			}
			if json.NewDecoder(resp.Body).Decode(&body) == nil {
				lastCount = len(body.Wants)
			}
			resp.Body.Close()
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
