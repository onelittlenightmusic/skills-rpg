package server

import (
	"encoding/json"
	"fmt"
	"strings"
)

// mywantBoard answers the rules' questions about named values by asking mywant.
//
// The catalog a subtype files into is not always the subtype's own name — a
// `city` is filed under `cities` — so both the reads and the writes go through
// the subtype and let mywant resolve it, rather than this file guessing at
// plurals.
type mywantBoard struct{ url string }

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

func (b mywantBoard) things() []mywantThing {
	body, err := mywantGet(b.url, "/api/v1/things")
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
func (b mywantBoard) find(subtype, value string) (mywantThing, bool) {
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

func (b mywantBoard) Named(subtype, value string) bool {
	_, ok := b.find(subtype, value)
	return ok
}

// Pinned is deliberately not "is it drawn on the canvas".
//
// A value is drawn when something refers to it OR when it has been pinned, and
// the fortress turns on telling those apart: the redaction crews work by
// removing the references, so a tile that is only drawn because a want happens
// to name it is exactly the tile they can take away. Only the label counts.
func (b mywantBoard) Pinned(subtype, value string) bool {
	t, ok := b.find(subtype, value)
	if !ok {
		return false
	}
	return t.Labels[thingPinLabel] == "true"
}

func (b mywantBoard) Name(subtype, value string) (string, error) {
	if t, ok := b.find(subtype, value); ok {
		return t.ID, nil // already told; saying it twice is not an error
	}
	body, err := mywantPost(b.url, "/api/v1/things", map[string]any{
		"catalog": subtype,
		"value":   value,
	})
	if err != nil {
		return "", fmt.Errorf("the city would not take the name: %w", err)
	}
	var out struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &out)
	return out.ID, nil
}

func (b mywantBoard) Pin(subtype, value string) error {
	t, ok := b.find(subtype, value)
	if !ok {
		return fmt.Errorf("nothing named %q to pin", value)
	}
	_, err := mywantPost(b.url, "/api/v1/things/"+t.ID+"/labels", map[string]any{
		"key":   thingPinLabel,
		"value": "true",
	})
	if err != nil {
		return fmt.Errorf("could not put %q on the board: %w", value, err)
	}
	return nil
}
