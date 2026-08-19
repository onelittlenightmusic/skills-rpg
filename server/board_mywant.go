package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// mywantBoard answers the rules' questions about named values by asking mywant.
//
// The catalog a subtype files into is not always the subtype's own name — a
// `city` is filed under `cities` — so both the reads and the writes go through
// the subtype and let mywant resolve it, rather than this file guessing at
// plurals.
type mywantBoard struct {
	url string
	mu  sync.Mutex
	// One pass over the activators asks the same four questions of mywant
	// twenty-odd times over. Every answer comes from one of four endpoints, so
	// the pass takes one copy of each and works from that.
	//
	// The window is short on purpose: this is not a cache of the city, it is the
	// snapshot a single reconciliation is reasoning about. Two passes a second
	// apart see different boards, which is what makes a change a change.
	cache map[string]cachedFetch
}

type cachedFetch struct {
	body []byte
	err  error
	at   time.Time
}

// How long a fetched endpoint stands in for the live one. Long enough to cover
// one sweep of the activators, short enough that the next sweep is fresh.
const boardSnapshotTTL = 750 * time.Millisecond

// fetch is every read of mywant this file makes, snapshotted.
func (b *mywantBoard) fetch(path string) ([]byte, error) {
	b.mu.Lock()
	if c, ok := b.cache[path]; ok && time.Since(c.at) < boardSnapshotTTL {
		b.mu.Unlock()
		return c.body, c.err
	}
	b.mu.Unlock()

	body, err := mywantGet(b.url, path)

	b.mu.Lock()
	if b.cache == nil {
		b.cache = map[string]cachedFetch{}
	}
	b.cache[path] = cachedFetch{body: body, err: err, at: time.Now()}
	b.mu.Unlock()
	return body, err
}

// The label mywant uses for the user's own decision that a value stands on the
// canvas. Absent means nobody has said either way and the automatic rule
// applies; "false" means it was taken off by hand. See the GUI's thingTileStore,
// which is the other reader of this label.
const thingPinLabel = "mywant.io/canvas"

type mywantThing struct {
	ID      string            `json:"id"`
	Catalog string            `json:"catalog"`
	Subtype string            `json:"subtype"`
	Value   string            `json:"value"`
	Labels  map[string]string `json:"labels"`
}

func (b *mywantBoard) things() []mywantThing {
	body, err := b.fetch("/api/v1/things")
	if err != nil {
		return nil
	}
	var out struct {
		Things []mywantThing `json:"things"`
	}
	if json.Unmarshal(body, &out) != nil {
		return nil
	}
	return out.Things
}

// find returns the thing matching this subtype and value, if the city holds one.
// Matching is case-insensitive on the value because the player types it, and
// tolerant on the subtype because a stage says `maker` while mywant may file it
// under `makers`.
func (b *mywantBoard) find(subtype, value string) (mywantThing, bool) {
	for _, t := range b.things() {
		if !strings.EqualFold(t.Value, value) {
			continue
		}
		if subtypeMatches(subtype, t) {
			return t, true
		}
	}
	return mywantThing{}, false
}

func subtypeMatches(subtype string, t mywantThing) bool {
	if subtype == "" {
		return true
	}
	return strings.EqualFold(t.Subtype, subtype) ||
		strings.EqualFold(t.Catalog, subtype) ||
		strings.EqualFold(t.Catalog, subtype+"s")
}

// CountNamed counts what the city holds under this subtype. An empty value is
// not "a value that happens to be blank" — it is "any", which is what a stage
// asks for when the point is that the player typed something, not what.
func (b *mywantBoard) CountNamed(subtype, value string) int {
	n := 0
	for _, t := range b.things() {
		if value != "" && !strings.EqualFold(t.Value, value) {
			continue
		}
		if subtypeMatches(subtype, t) {
			n++
		}
	}
	return n
}

// Pinned is deliberately not "is it drawn on the canvas".
//
// A value is drawn when something refers to it OR when it has been pinned, and
// the fortress turns on telling those apart: the redaction crews work by
// removing the references, so a tile that is only drawn because a want happens
// to name it is exactly the tile they can take away. Only the label counts.
func (b *mywantBoard) Pinned(subtype, value string) bool {
	t, ok := b.find(subtype, value)
	if !ok {
		return false
	}
	return t.Labels[thingPinLabel] == "true"
}

// UsedBy counts the live wants naming this value.
//
// Read from mywant's own usage relation rather than counted here: it is derived
// there from each want's subtyped parameters, and a second implementation would
// be a second opinion about what "uses" means — the tile on the board draws its
// roads from that same relation, so the number the player sees and the number a
// door waits for have to come from one place.
func (b *mywantBoard) UsedBy(subtype, value string) int {
	body, err := b.fetch("/api/v1/memo/usage")
	if err != nil {
		return 0
	}
	var out struct {
		Usage []struct {
			Catalog string   `json:"catalog"`
			Subtype string   `json:"subtype"`
			Value   string   `json:"value"`
			WantIDs []string `json:"wantIDs"`
		} `json:"usage"`
	}
	if json.Unmarshal(body, &out) != nil {
		return 0
	}
	for _, u := range out.Usage {
		if !strings.EqualFold(u.Value, value) {
			continue
		}
		if subtype == "" || strings.EqualFold(u.Subtype, subtype) ||
			strings.EqualFold(u.Catalog, subtype) || strings.EqualFold(u.Catalog, subtype+"s") {
			return len(u.WantIDs)
		}
	}
	return 0
}

// GroupMembers counts what has been gathered under one name.
//
// mywant calls these constellations and serves them at /api/v1/groups; a group
// is a name and a member list, which is exactly the shape fortress7 needs — the
// ward ledger will not take eight buildings, it takes one district.
func (b *mywantBoard) GroupMembers(name string) int {
	body, err := b.fetch("/api/v1/groups")
	if err != nil {
		return 0
	}
	var out struct {
		Groups []struct {
			ID      string   `json:"id"`
			Name    string   `json:"name"`
			Members []string `json:"members"`
		} `json:"groups"`
	}
	if json.Unmarshal(body, &out) != nil {
		return 0
	}
	for _, g := range out.Groups {
		if strings.EqualFold(g.Name, name) || strings.EqualFold(g.ID, name) {
			return len(g.Members)
		}
	}
	return 0
}

type mywantWant struct {
	Metadata struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"metadata"`
	Status string `json:"status"`
	State  struct {
		Current map[string]any `json:"current"`
		Goal    map[string]any `json:"goal"`
		Plan    map[string]any `json:"plan"`
	} `json:"state"`
}

func (b *mywantBoard) wants() []mywantWant {
	body, err := b.fetch("/api/v1/wants")
	if err != nil {
		return nil
	}
	var out struct {
		Wants []mywantWant `json:"wants"`
	}
	if json.Unmarshal(body, &out) != nil {
		return nil
	}
	return out.Wants
}

// CountWants counts wants by type, name and state. Empty fields match anything,
// so a stage can ask "how many of these are running" or "is that one still
// there" with the same condition.
func (b *mywantBoard) CountWants(wantType, name, state string) int {
	n := 0
	for _, w := range b.wants() {
		if wantType != "" && !strings.EqualFold(w.Metadata.Type, wantType) {
			continue
		}
		if name != "" && !strings.EqualFold(w.Metadata.Name, name) && !strings.EqualFold(w.Metadata.ID, name) {
			continue
		}
		if state != "" && !strings.EqualFold(w.Status, state) {
			continue
		}
		n++
	}
	return n
}

// WantState reads one field out of a want's current state.
//
// The state, not the spec: what a want has actually produced is the question
// Act 2 turns on, and a parameter is what it was asked for. The two agree right
// up until the moment the stage is about.
func (b *mywantBoard) WantState(name, field string) string {
	for _, w := range b.wants() {
		if !strings.EqualFold(w.Metadata.Name, name) && !strings.EqualFold(w.Metadata.ID, name) {
			continue
		}
		for _, bucket := range []map[string]any{w.State.Current, w.State.Goal, w.State.Plan} {
			if v, ok := bucket[field]; ok {
				return fmt.Sprint(v)
			}
		}
	}
	return ""
}

// Connected reports whether a road runs from one want to another.
//
// A road in mywant is a relation: one want exposes a field and another imports
// it. Read from /api/v1/relations rather than inferred from either want's own
// metadata, for the same reason UsedBy is — the line the player draws on the
// canvas and the line a door waits for have to be the same line.
func (b *mywantBoard) Connected(from, to, field string) bool {
	is := func(want, id, name string) bool {
		return want == "" || strings.EqualFold(want, id) || strings.EqualFold(want, name)
	}
	for _, r := range b.relations() {
		if !is(from, r.ProviderID, r.ProviderName) || !is(to, r.ConsumerID, r.ConsumerName) {
			continue
		}
		if field != "" && !strings.EqualFold(field, r.FieldName) {
			continue
		}
		return true
	}
	return false
}

// relations is every road on the board, as mywant reports them.
func (b *mywantBoard) relations() []mywantRelation {
	body, err := b.fetch("/api/v1/relations")
	if err != nil {
		return nil
	}
	var out struct {
		Relations []mywantRelation `json:"relations"`
	}
	if json.Unmarshal(body, &out) != nil {
		return nil
	}
	return out.Relations
}

type mywantRelation struct {
	ProviderID   string `json:"provider_id"`
	ProviderName string `json:"provider_name"`
	ConsumerID   string `json:"consumer_id"`
	ConsumerName string `json:"consumer_name"`
	FieldName    string `json:"field_name"`
}

// ConnectionsFrom counts the roads leaving a want — one per want it feeds, which
// is the number the finale asks the player to get to eight.
func (b *mywantBoard) ConnectionsFrom(name string) int {
	seen := map[string]bool{}
	for _, r := range b.relations() {
		if !strings.EqualFold(name, r.ProviderID) && !strings.EqualFold(name, r.ProviderName) {
			continue
		}
		seen[r.ConsumerID] = true
	}
	return len(seen)
}
