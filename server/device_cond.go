package server

import (
	"encoding/json"
	"fmt"
	"strings"
)

// What an activator reads off the city, and what it says about it.
//
// World 1's devices are switches: chap flips one and a door takes power. World
// 2's are readers — they run while the city is in some state, and the player's
// job is to put it in that state. The door in front of either says only
// `requires_device`, because a door is a door.
//
// A reader is useless if the player cannot tell what it is waiting for, and
// "off" says nothing. So every condition answers three questions and not one:
//
//	satisfied() - is it running
//	checks()    - the parts of the question, each true or false on its own
//	need()      - the one job that is outstanding right now, in the imperative
//
// checks() and need() are here, in Go, next to the condition itself, rather
// than reconstructed by the skill that draws the card: the thing that owns the
// fact answers for it. Adding a condition kind below is all it takes for its
// card to explain itself — no Python, no JSX.

// Check is one part of a condition, as the card lists it.
type Check struct {
	Label string `json:"label"`
	OK    bool   `json:"ok"`
}

// deviceCond is a question a device asks the city.
type deviceCond interface {
	satisfied() bool
	checks() []Check
	need() string
	// subject is the one noun this is about, for a narration to name.
	subject() string
}

// ---------------------------------------------------------------- named value

// The value a condition is about is written in the stage, plainly. An earlier
// version let a door defer to an item instead — `from_item: lira_mark` — so
// that the card would not give away a name the player was meant to go and read.
// What it gave away instead was an internal id: "name the person on lira_mark"
// means nothing to somebody seeing this for the first time, and no wording
// rescued it. A condition that says which name is missing can be understood; one
// that says which item holds it cannot.

// ThingCond names a value the city has to know about — and, if Pinned, one the
// player has put on the board on their own account rather than one that happens
// to be drawn because something refers to it. An empty Value asks for any value
// of that kind, which is how a stage asks the player to type one in without
// dictating what.
type ThingCond struct {
	Subtype string `yaml:"subtype,omitempty" json:"subtype,omitempty"`
	Value   string `yaml:"value,omitempty" json:"value,omitempty"`
	Pinned  bool   `yaml:"pinned,omitempty" json:"pinned,omitempty"`
}

// Want returns the value this condition is about.
func (c *ThingCond) Want(*Stage) string {
	if c == nil {
		return ""
	}
	return c.Value
}

func (c *ThingCond) named() bool { return currentBoard.CountNamed(c.Subtype, c.Value) > 0 }

func (c *ThingCond) satisfied() bool {
	if c.Pinned {
		return currentBoard.Pinned(c.Subtype, c.Value)
	}
	return c.named()
}

func (c *ThingCond) subject() string {
	if c.Value == "" {
		return "any " + c.Subtype
	}
	return c.Value
}

func (c *ThingCond) label() string {
	if c.Value == "" {
		return "any " + c.Subtype
	}
	return fmt.Sprintf("%s: %s", c.Subtype, c.Value)
}

func (c *ThingCond) checks() []Check {
	out := []Check{{Label: c.label(), OK: c.named()}}
	if c.Pinned {
		out = append(out, Check{Label: "pinned", OK: currentBoard.Pinned(c.Subtype, c.Value)})
	}
	return out
}

// A value the city has never heard of cannot be pinned, so the two halves are
// two jobs in an order, and only one of them is the next one.
func (c *ThingCond) need() string {
	if c.satisfied() {
		return ""
	}
	if !c.named() {
		if c.Value == "" {
			return fmt.Sprintf("The registry holds no %s. Give the city one — any value will do.", c.Subtype)
		}
		return fmt.Sprintf("No %s called %q in the registry. Add one.", c.Subtype, c.Value)
	}
	if c.Pinned {
		return fmt.Sprintf("%s is in the registry but not pinned. Pin it, so it stands on the board on its own account.", c.Value)
	}
	return ""
}

// ----------------------------------------------------------------- roads out

// UsageCond makes a device a meter on the roads leaving a thing tile: it runs
// once at least MinUsers live wants name the value. The board draws one road
// per user, so what the meter reads is what the player can count.
type UsageCond struct {
	Subtype  string `yaml:"subtype,omitempty" json:"subtype,omitempty"`
	Value    string `yaml:"value,omitempty" json:"value,omitempty"`
	MinUsers int    `yaml:"min_users,omitempty" json:"min_users,omitempty"`
}

func (c *UsageCond) min() int        { return max(c.MinUsers, 1) }
func (c *UsageCond) count() int      { return currentBoard.UsedBy(c.Subtype, c.Value) }
func (c *UsageCond) subject() string { return c.Value }

func (c *UsageCond) satisfied() bool { return c.count() >= c.min() }

func (c *UsageCond) checks() []Check {
	return []Check{{
		Label: fmt.Sprintf("%s: used by %d of %d", c.Value, c.count(), c.min()),
		OK:    c.satisfied(),
	}}
}

func (c *UsageCond) need() string {
	short := c.min() - c.count()
	if short <= 0 {
		return ""
	}
	if currentBoard.CountNamed(c.Subtype, c.Value) == 0 {
		return fmt.Sprintf("No %s called %q in the registry. Add one, then give it to a want.", c.Subtype, c.Value)
	}
	return fmt.Sprintf("%s is used by %d want(s); this needs %d. Put the value into %d more want(s) — every user draws a road to the tile.",
		c.Value, c.count(), c.min(), short)
}

// -------------------------------------------------------------------- a group

// GroupCond runs once a group of the given name holds MinMembers things. One
// name for many values, which is the first relationship the player names rather
// than a value.
type GroupCond struct {
	Name       string `yaml:"name,omitempty" json:"name,omitempty"`
	MinMembers int    `yaml:"min_members,omitempty" json:"min_members,omitempty"`
}

func (c *GroupCond) min() int        { return max(c.MinMembers, 1) }
func (c *GroupCond) count() int      { return currentBoard.GroupMembers(c.Name) }
func (c *GroupCond) subject() string { return c.Name }

func (c *GroupCond) satisfied() bool { return c.count() >= c.min() }

func (c *GroupCond) checks() []Check {
	return []Check{{
		Label: fmt.Sprintf("group %s: %d of %d", c.Name, c.count(), c.min()),
		OK:    c.satisfied(),
	}}
}

func (c *GroupCond) need() string {
	short := c.min() - c.count()
	if short <= 0 {
		return ""
	}
	if c.count() == 0 {
		return fmt.Sprintf("There is no group called %q. Select %d things on the board and group them under that name.", c.Name, c.min())
	}
	return fmt.Sprintf("The group %q holds %d of %d. Add %d more thing(s) to it.", c.Name, c.count(), c.min(), short)
}

// --------------------------------------------------------------------- a want

// WantCond runs on the presence — or, with Absent, the departure — of wants
// matching a description. State is the status a want has to be in; empty means
// any. Name matches the want's name, Type its type; either may be empty.
type WantCond struct {
	Type     string `yaml:"type,omitempty" json:"type,omitempty"`
	Name     string `yaml:"name,omitempty" json:"name,omitempty"`
	State    string `yaml:"state,omitempty" json:"state,omitempty"`
	MinCount int    `yaml:"min_count,omitempty" json:"min_count,omitempty"`
	// Absent flips the question: the device runs while NOTHING matches. For the
	// stage whose job is to stop something, not to build something.
	Absent bool `yaml:"absent,omitempty" json:"absent,omitempty"`
}

func (c *WantCond) min() int   { return max(c.MinCount, 1) }
func (c *WantCond) count() int { return currentBoard.CountWants(c.Type, c.Name, c.State) }

func (c *WantCond) satisfied() bool {
	if c.Absent {
		return c.count() == 0
	}
	return c.count() >= c.min()
}

// describe names what is being looked for, in the words the board uses.
func (c *WantCond) describe() string {
	parts := []string{}
	if c.State != "" {
		parts = append(parts, c.State)
	}
	if c.Type != "" {
		parts = append(parts, c.Type)
	}
	out := "want"
	if len(parts) > 0 {
		out = strings.Join(parts, " ") + " want"
	}
	if c.Name != "" {
		out += " " + c.Name
	}
	return out
}

func (c *WantCond) subject() string {
	if c.Name != "" {
		return c.Name
	}
	return c.describe()
}

func (c *WantCond) checks() []Check {
	label := fmt.Sprintf("%s: %d", c.describe(), c.count())
	if c.Absent {
		label = fmt.Sprintf("%s: none (%d)", c.describe(), c.count())
	}
	return []Check{{Label: label, OK: c.satisfied()}}
}

func (c *WantCond) need() string {
	if c.satisfied() {
		return ""
	}
	if c.Absent {
		if c.Name != "" && c.State != "" {
			return fmt.Sprintf("The want %s is %s, and this is waiting for it not to be. Change that.", c.Name, c.State)
		}
		if c.Name != "" {
			return fmt.Sprintf("The want %s is still on the board. Stop it — this reads the board, and that is what it is still seeing.", c.Name)
		}
		if c.State != "" {
			return fmt.Sprintf("%d want(s) on the board are still %s, and this is waiting for none. Find them and change that.", c.count(), c.State)
		}
		return fmt.Sprintf("There is still a %s on the board. Stop it — this reads the board, and that is what it is still seeing.", c.describe())
	}
	short := c.min() - c.count()
	if c.Name != "" && c.State != "" {
		return fmt.Sprintf("The want %s is not %s. Open its card and get it %s.", c.Name, c.State, c.State)
	}
	return fmt.Sprintf("This needs %d %s and there are %d. Deploy %d more.", c.min(), c.describe(), c.count(), short)
}

// ------------------------------------------------------------ what it has made

// StateCond reads a value a want is holding — what it has actually produced,
// rather than whether it is running. Act 2's distinction, and the reason a row
// of green wants can still be wrong.
type StateCond struct {
	Name   string `yaml:"name,omitempty" json:"name,omitempty"`
	Field  string `yaml:"field,omitempty" json:"field,omitempty"`
	Equals string `yaml:"equals,omitempty" json:"equals,omitempty"`
}

func (c *StateCond) actual() string { return currentBoard.WantState(c.Name, c.Field) }

func (c *StateCond) satisfied() bool {
	got := c.actual()
	if c.Equals == "" {
		return got != ""
	}
	return strings.EqualFold(got, c.Equals)
}

func (c *StateCond) subject() string { return c.Name + "." + c.Field }

func (c *StateCond) checks() []Check {
	got := c.actual()
	if got == "" {
		got = "—"
	}
	return []Check{{Label: fmt.Sprintf("%s.%s = %s", c.Name, c.Field, got), OK: c.satisfied()}}
}

func (c *StateCond) need() string {
	if c.satisfied() {
		return ""
	}
	got := c.actual()
	if got == "" {
		return fmt.Sprintf("%s is not producing a %s yet. Open its card and give it one.", c.Name, c.Field)
	}
	return fmt.Sprintf("%s is putting out %s = %s, and this wants %s. Correct it on its card.",
		c.Name, c.Field, got, c.Equals)
}

// --------------------------------------------------------------------- a road

// ConnCond runs while a road runs from one want to another — mywant's own
// relation, the same line the board draws between two tiles. Field names the
// value carried, when the stage is about which value and not merely whether.
type ConnCond struct {
	From  string `yaml:"from,omitempty" json:"from,omitempty"`
	To    string `yaml:"to,omitempty" json:"to,omitempty"`
	Field string `yaml:"field,omitempty" json:"field,omitempty"`
	// MinCount asks for a number of roads leaving From rather than for one
	// particular road. The finale's question: eight buildings fed from one
	// supply, and the difference between wiring them and saying they are a set.
	MinCount int `yaml:"min_count,omitempty" json:"min_count,omitempty"`
	// Absent flips the question — this road must NOT be there. For the stage
	// where a plausible connection is the wrong one, and taking it out is the
	// move.
	Absent bool `yaml:"absent,omitempty" json:"absent,omitempty"`
	// Chain is a road with stops: every consecutive pair has to be joined. For
	// the stage whose point is that no single want holds the value the door
	// wants — one produces, one transforms, one delivers.
	Chain []string `yaml:"chain,omitempty" json:"chain,omitempty"`
}

// hops is the pairs that have to be joined, chain or not.
func (c *ConnCond) hops() [][2]string {
	if len(c.Chain) >= 2 {
		out := make([][2]string, 0, len(c.Chain)-1)
		for i := 0; i+1 < len(c.Chain); i++ {
			out = append(out, [2]string{c.Chain[i], c.Chain[i+1]})
		}
		return out
	}
	return [][2]string{{c.From, c.To}}
}

func (c *ConnCond) fanOut() int { return currentBoard.ConnectionsFrom(c.From) }

func (c *ConnCond) satisfied() bool {
	if c.MinCount > 0 {
		return c.fanOut() >= c.MinCount
	}
	if c.Absent {
		return !currentBoard.Connected(c.From, c.To, c.Field)
	}
	for _, h := range c.hops() {
		if !currentBoard.Connected(h[0], h[1], c.Field) {
			return false
		}
	}
	return true
}

func (c *ConnCond) subject() string {
	if len(c.Chain) >= 2 {
		return strings.Join(c.Chain, " → ")
	}
	return c.From + " → " + c.To
}

func (c *ConnCond) checks() []Check {
	if c.MinCount > 0 {
		return []Check{{
			Label: fmt.Sprintf("%s: %d of %d roads", c.From, c.fanOut(), c.MinCount),
			OK:    c.satisfied(),
		}}
	}
	if c.Absent {
		return []Check{{Label: fmt.Sprintf("no %s → %s", c.From, c.To), OK: c.satisfied()}}
	}
	var out []Check
	for _, h := range c.hops() {
		label := fmt.Sprintf("%s → %s", h[0], h[1])
		if c.Field != "" {
			label += " (" + c.Field + ")"
		}
		out = append(out, Check{Label: label, OK: currentBoard.Connected(h[0], h[1], c.Field)})
	}
	return out
}

// The first missing hop, and only that one. A chain is built one road at a time
// and being told about all of them at once is being told about none of them.
func (c *ConnCond) need() string {
	if c.MinCount > 0 {
		if c.satisfied() {
			return ""
		}
		return fmt.Sprintf("%d road(s) leave %s and this needs %d. Draw %d more.",
			c.fanOut(), c.From, c.MinCount, c.MinCount-c.fanOut())
	}
	if c.Absent {
		if c.satisfied() {
			return ""
		}
		return fmt.Sprintf("%s is still joined to %s, and it should not be. Disconnect them.", c.From, c.To)
	}
	for _, h := range c.hops() {
		if currentBoard.Connected(h[0], h[1], c.Field) {
			continue
		}
		what := ""
		if c.Field != "" {
			what = fmt.Sprintf(" It is %s that has to travel.", c.Field)
		}
		return fmt.Sprintf("Nothing joins %s to %s. Connect them: %s offers the value, %s asks for it.%s",
			h[0], h[1], h[0], h[1], what)
	}
	return ""
}

// ---------------------------------------------------------------- the device

// cond returns the one question this device asks the city, or nil for a plain
// switch. Stages declare at most one; the order here is only a tiebreak.
func (d *Device) cond() deviceCond {
	switch {
	case d == nil:
		return nil
	case d.Reads != nil:
		return d.Reads
	case d.ReadsUsed != nil:
		return d.ReadsUsed
	case d.ReadsGroup != nil:
		return d.ReadsGroup
	case d.ReadsWant != nil:
		return d.ReadsWant
	case d.ReadsState != nil:
		return d.ReadsState
	case d.ReadsConn != nil:
		return d.ReadsConn
	}
	return nil
}

// IsOn reports whether the device is running: a reader answers from the city,
// everything else keeps the switch chap flips.
func (d *Device) IsOn() bool {
	if d == nil {
		return false
	}
	if c := d.cond(); c != nil {
		return c.satisfied()
	}
	return d.On
}

// Need is the one job outstanding, in the imperative, or "" when there is none.
// The stage's own NeedHint follows it, for the concrete how-to that only the
// stage knows.
func (d *Device) Need() string {
	c := d.cond()
	if c == nil {
		return ""
	}
	need := c.need()
	if need == "" {
		return ""
	}
	if d.NeedHint != "" {
		return need + " " + d.NeedHint
	}
	return need
}

// Checks lists the parts of the condition, each answered on its own.
func (d *Device) Checks() []Check {
	if c := d.cond(); c != nil {
		return c.checks()
	}
	return nil
}

// MarshalJSON reports the device as it actually stands — clients read `on` out
// of the state, and a reader's answer has to be what they get — together with
// what it is waiting for, so a card can say so without knowing the kinds.
func (d Device) MarshalJSON() ([]byte, error) {
	type plain Device
	out := struct {
		plain
		On     bool    `json:"on"`
		Checks []Check `json:"checks,omitempty"`
		Need   string  `json:"need,omitempty"`
	}{plain: plain(d), On: d.IsOn(), Checks: d.Checks(), Need: d.Need()}
	return json.Marshal(out)
}

// Subject is what a change to this device is about, for the narration that
// reports it. Empty for a plain switch, which has nothing to be about.
func (d *Device) Subject() string {
	if c := d.cond(); c != nil {
		return c.subject()
	}
	return ""
}

// IsReader reports whether this device answers from the city rather than from a
// switch — the thing that decides how it is drawn, and whether anybody in the
// game can operate it.
func (d *Device) IsReader() bool { return d.cond() != nil }
