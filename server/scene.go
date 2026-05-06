package server

import "sort"

// SceneNode represents a waypoint/room for rendering.
type SceneNode struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	HasYou bool   `json:"has_you"`
}

// SceneDoor is the door connecting two nodes.
type SceneDoor struct {
	ID     string `json:"id"`
	Open   bool   `json:"open"`
	Locked bool   `json:"locked"`
}

// SceneEdge connects two nodes, optionally via a door.
type SceneEdge struct {
	From string     `json:"from"`
	To   string     `json:"to"`
	Door *SceneDoor `json:"door,omitempty"`
}

// SceneDevice represents a device in the current stage.
type SceneDevice struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	On    bool   `json:"on"`
}

// Scene is the full render descriptor for the current stage.
type Scene struct {
	StageID      string        `json:"stage_id"`
	Title        string        `json:"title"`
	Nodes        []SceneNode   `json:"nodes"`
	Edges        []SceneEdge   `json:"edges"`
	Devices      []SceneDevice `json:"devices"`
	NextGoal     string        `json:"next_goal"`
	EventHistory []Event       `json:"event_history,omitempty"`
	ChapItems    []string      `json:"chap_items,omitempty"`
}

// BuildScene computes a Scene from the current game state with locale overlay applied.
func (s *Server) BuildScene() Scene {
	s.mu.Lock()
	gs := cloneState(s.state)
	locale := s.currentLocaleLocked()
	s.mu.Unlock()

	stageID := gs.CurrentStage
	stage := gs.Stages[stageID]
	if stage == nil {
		return Scene{StageID: stageID}
	}

	youPos := gs.You.Position

	// BFS order of waypoints, starting from initial_position for stable ordering
	chain := bfsWaypoints(stage.Waypoints, stage.InitialPosition)

	nodes := make([]SceneNode, 0, len(chain))
	for _, id := range chain {
		wp := stage.Waypoints[id]
		label := id
		if wp != nil && wp.Label != "" {
			label = wp.Label
		}
		if locale != nil {
			if wl, ok := locale.Waypoints[id]; ok && wl.Label != "" {
				label = wl.Label
			}
		}
		nodes = append(nodes, SceneNode{
			ID:     id,
			Label:  label,
			HasYou: id == youPos,
		})
	}

	// Edges: consecutive pairs in BFS chain
	edges := make([]SceneEdge, 0)
	for i := 1; i < len(chain); i++ {
		a, b := chain[i-1], chain[i]
		edge := SceneEdge{From: a, To: b}
		for doorID, d := range stage.Doors {
			if (d.Between[0] == a && d.Between[1] == b) ||
				(d.Between[0] == b && d.Between[1] == a) {
				edge.Door = &SceneDoor{
					ID:     doorID,
					Open:   d.Open,
					Locked: d.Locked,
				}
				break
			}
		}
		edges = append(edges, edge)
	}

	devices := make([]SceneDevice, 0, len(stage.Devices))
	for id, d := range stage.Devices {
		dlabel := d.Label
		if locale != nil {
			if dl, ok := locale.Devices[id]; ok && dl.Label != "" {
				dlabel = dl.Label
			}
		}
		devices = append(devices, SceneDevice{
			ID:    id,
			Label: dlabel,
			On:    d.On,
		})
	}

	var chapItems []string
	for id, item := range stage.Items {
		if item.HeldBy == ActorChap {
			chapItems = append(chapItems, id)
		}
	}
	sort.Strings(chapItems)

	title := stage.Title
	if locale != nil && locale.Title != "" {
		title = locale.Title
	}

	nextGoalText := computeNextGoal(gs, locale).Text

	return Scene{
		StageID:      stageID,
		Title:        title,
		Nodes:        nodes,
		Edges:        edges,
		Devices:      devices,
		NextGoal:     nextGoalText,
		EventHistory: gs.EventHistory,
		ChapItems:    chapItems,
	}
}

// bfsWaypoints returns waypoint IDs in BFS order starting from initialPos.
func bfsWaypoints(wps map[string]*Waypoint, initialPos string) []string {
	if len(wps) == 0 {
		return nil
	}
	start := initialPos
	if start == "" || wps[start] == nil {
		// fallback: alphabetically first key for determinism
		keys := make([]string, 0, len(wps))
		for k := range wps {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		start = keys[0]
	}
	out := []string{}
	seen := map[string]bool{start: true}
	q := []string{start}
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		out = append(out, cur)
		if wps[cur] != nil {
			for _, nb := range wps[cur].Adjacent {
				if !seen[nb] {
					seen[nb] = true
					q = append(q, nb)
				}
			}
		}
	}
	for k := range wps {
		if !seen[k] {
			out = append(out, k)
		}
	}
	return out
}
