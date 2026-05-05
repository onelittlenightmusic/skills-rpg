package server

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

// BuildScene computes a Scene from the current game state.
func (s *Server) BuildScene() Scene {
	gs := s.State()
	stageID := gs.CurrentStage
	stage := gs.Stages[stageID]
	if stage == nil {
		return Scene{StageID: stageID}
	}

	youPos := gs.You.Position

	// BFS order of waypoints
	chain := bfsWaypoints(stage.Waypoints)

	nodes := make([]SceneNode, 0, len(chain))
	for _, id := range chain {
		wp := stage.Waypoints[id]
		label := id
		if wp != nil && wp.Label != "" {
			label = wp.Label
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
		devices = append(devices, SceneDevice{
			ID:    id,
			Label: d.Label,
			On:    d.On,
		})
	}

	var chapItems []string
	for id, item := range stage.Items {
		if item.HeldBy == ActorChap {
			chapItems = append(chapItems, id)
		}
	}

	return Scene{
		StageID:      stageID,
		Title:        stage.Title,
		Nodes:        nodes,
		Edges:        edges,
		Devices:      devices,
		NextGoal:     gs.NextGoal.Text,
		EventHistory: gs.EventHistory,
		ChapItems:    chapItems,
	}
}

// bfsWaypoints returns waypoint IDs in BFS order.
func bfsWaypoints(wps map[string]*Waypoint) []string {
	if len(wps) == 0 {
		return nil
	}
	// deterministic start: first key in iteration (stable enough for display)
	var start string
	for k := range wps {
		start = k
		break
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
