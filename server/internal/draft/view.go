package draft

import (
	"time"

	"github.com/google/uuid"
)

// View is a draft as one seat may see it: the JSON payload of a `draft`
// frame (docs/protocol.md) and of the HTTP answers that return a draft.
//
// # Hidden information
//
// CR 905.1c: "During the draft, a player can look only at cards in the
// booster pack they are currently drafting from, cards they have
// already drafted, ..." A View therefore holds cards in three fields
// only (Pack, Picks, Pool), and all three are the viewer's own. Every
// other seat appears as a PlayerView, which has no card field at all: a
// name, whether it is still picking, and how many cards it has drafted.
// There is no unfiltered view to redact, so nothing can leak by a
// filter being forgotten. TestNoSeatSeesAnotherSeatsCards walks every
// view of a whole draft to hold this.
type View struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Set     string    `json:"set"`
	SetName string    `json:"set_name"`
	Booster string    `json:"booster,omitempty"`
	Mode    Mode      `json:"mode"`
	Seats   int       `json:"seats"`
	Solo    bool      `json:"solo"`
	State   State     `json:"state"`
	Seq     uint64    `json:"seq"`

	Round        int `json:"round"`
	Rounds       int `json:"rounds"`
	Pick         int `json:"pick"`
	PicksInRound int `json:"picks_in_round"`
	// Take is how many cards the viewer must take from the pack it
	// holds now, 0 when it holds none.
	Take    int       `json:"take"`
	Passing Direction `json:"passing,omitempty"`

	PickSeconds int        `json:"pick_seconds"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Now         *time.Time `json:"now,omitempty"`

	// You is the viewer's seat number.
	You     int          `json:"you"`
	Players []PlayerView `json:"players"`

	// Pack is the pack the viewer holds, absent once it has picked.
	Pack []Card `json:"pack,omitempty"`
	// Picks is the viewer's own picks in the order taken.
	Picks []Card `json:"picks"`
	// Pool is the viewer's card pool, present once the draft is done.
	Pool []Card `json:"pool,omitempty"`

	// Invite is the invite code, set by the lobby for the host of a
	// draft still in its lobby. The draft itself never fills it.
	Invite string `json:"invite,omitempty"`
}

// PlayerView is one seat as every seat may see it. It has no card field
// on purpose (see View).
type PlayerView struct {
	Seat      int    `json:"seat"`
	Name      string `json:"name"`
	Bot       bool   `json:"bot,omitempty"`
	Host      bool   `json:"host,omitempty"`
	Connected bool   `json:"connected"`
	// Picking is true while the seat has not yet picked from the pack
	// it holds.
	Picking bool `json:"picking"`
	// Picked is how many cards the seat has drafted.
	Picked int `json:"picked"`
}

// ViewFor builds the view of one seat. connected reports whether a seat
// has a live socket; a bot is always connected. ok is false for a seat
// the draft does not have.
func (d *Draft) ViewFor(seatID uuid.UUID, now time.Time, connected func(uuid.UUID) bool) (View, bool) {
	you, me := d.seat(seatID)
	if me == nil {
		return View{}, false
	}
	v := View{
		ID:           d.ID,
		Name:         d.cfg.Name,
		Set:          d.cfg.Set,
		SetName:      d.cfg.SetName,
		Booster:      d.cfg.Booster,
		Mode:         d.cfg.Mode,
		Seats:        d.cfg.Seats,
		Solo:         d.cfg.Solo,
		State:        d.state,
		Seq:          d.seq,
		Round:        d.round,
		Rounds:       Rounds,
		Pick:         d.pick,
		PicksInRound: d.picksInRound,
		PickSeconds:  int(d.cfg.PickTimeout / time.Second),
		You:          you,
		Players:      make([]PlayerView, 0, len(d.seats)),
		Picks:        append([]Card{}, me.picks...),
	}
	for i, s := range d.seats {
		v.Players = append(v.Players, PlayerView{
			Seat:      i,
			Name:      s.Name,
			Bot:       s.Bot,
			Host:      s.ID == d.host,
			Connected: s.Bot || (connected != nil && connected(s.ID)),
			Picking:   d.state == StatePicking && !s.picked,
			Picked:    len(s.picks),
		})
	}
	switch d.state {
	case StatePicking:
		v.Passing = DirectionOf(d.round)
		deadline, at := d.deadline, now
		v.Deadline, v.Now = &deadline, &at
		if !me.picked {
			v.Pack = append([]Card(nil), me.pack...)
			v.Take = d.take(len(me.pack))
		}
	case StateDone:
		v.Pool = append([]Card{}, me.picks...)
	}
	return v, true
}

// Preview is what someone holding the invite sees before taking a seat.
type Preview struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Set     string    `json:"set"`
	SetName string    `json:"set_name"`
	Mode    Mode      `json:"mode"`
	Seats   int       `json:"seats"`
	Taken   int       `json:"taken"`
	State   State     `json:"state"`
	Host    string    `json:"host,omitempty"`
}

// Preview builds the invite page's summary.
func (d *Draft) Preview() Preview {
	p := Preview{
		ID: d.ID, Name: d.cfg.Name, Set: d.cfg.Set, SetName: d.cfg.SetName,
		Mode: d.cfg.Mode, Seats: d.cfg.Seats, State: d.state,
	}
	for _, s := range d.seats {
		if !s.Bot {
			p.Taken++
		}
		if s.ID == d.host {
			p.Host = s.Name
		}
	}
	return p
}
