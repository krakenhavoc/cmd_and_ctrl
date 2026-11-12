package draft

import (
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
	// Cube is true for a draft of a saved card list (ADR 0222 §8): Set
	// is then empty and SetName is the cube's name.
	Cube  bool   `json:"cube,omitempty"`
	Mode  Mode   `json:"mode"`
	Seats int    `json:"seats"`
	Solo  bool   `json:"solo"`
	State State  `json:"state"`
	Seq   uint64 `json:"seq"`

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

	// Pack is the pack the viewer holds, absent once it has picked. A
	// card of Pack, Picks and Pool carries a `rating` (ratings.go).
	Pack []Card `json:"pack,omitempty"`
	// Picks is the viewer's own picks in the order taken.
	Picks []Card `json:"picks"`
	// Pool is the viewer's card pool, present once the draft is done.
	Pool []Card `json:"pool,omitempty"`

	// What follows the picks (limited.go, #5440). None of it holds a
	// card of another seat: Deck names the viewer's own pool cards by
	// ID, and a match names seats by number.
	//
	// Phase is where the draft is once State is done: building, rounds
	// or finished.
	Phase Phase `json:"phase,omitempty"`
	// BestOf is the length of the draft's matches, 1 or 3.
	BestOf int `json:"best_of"`
	// Build is the deckbuilding step, present while it is open.
	Build *BuildView `json:"build,omitempty"`
	// Deck is the viewer's own deck, once it has one.
	Deck *LimitedDeck `json:"deck,omitempty"`
	// Play is the rounds, present once they have begun.
	Play *PlayView `json:"play,omitempty"`

	// Invite is the invite code, set by the lobby for the host of a
	// draft still in its lobby. The draft itself never fills it.
	Invite string `json:"invite,omitempty"`
}

// BuildView is the deckbuilding step.
type BuildView struct {
	// Deadline is when the timer runs out. It is absent when the build
	// is untimed (#5700): the rounds begin when the last person has
	// submitted.
	Deadline *time.Time `json:"deadline,omitempty"`
	Now      time.Time  `json:"now"`
	// Seconds is the whole build timer, 0 when the build is untimed.
	Seconds int `json:"seconds"`
	// Untimed says the build has no timer.
	Untimed bool `json:"untimed,omitempty"`
	// DeckMin is the smallest legal deck (CR 100.2b).
	DeckMin int `json:"deck_min"`
	// Submitted says the viewer has submitted a deck. They may submit
	// again until the rounds begin.
	Submitted bool `json:"submitted"`
}

// PlayView is the rounds: the round in play, every round's matches and
// the win-loss table.
type PlayView struct {
	// Round is the round in play from 1, and Rounds how many a draft
	// plays.
	Round  int `json:"round"`
	Rounds int `json:"rounds"`
	// Matches is every round's matches so far, oldest first.
	Matches [][]MatchView `json:"matches"`
	// Standings is every seat's record, by seat number.
	Standings []Standing `json:"standings"`
	// Yours is the viewer's match in the round in play, absent when it
	// has none (the draft is finished).
	Yours *MatchView `json:"yours,omitempty"`
	// Table is the table the viewer's match is played at, sent to the
	// two seats of that match only and absent until it is open.
	Table *uuid.UUID `json:"table,omitempty"`
}

// MatchView is one match of a round, as every seat may see it.
type MatchView struct {
	// A and B are seat numbers. B is absent for a bye.
	A   int  `json:"a"`
	B   *int `json:"b,omitempty"`
	Bye bool `json:"bye,omitempty"`
	// Done says the match has a result; Winner is the seat that won it,
	// absent for a draw.
	Done   bool `json:"done"`
	Winner *int `json:"winner,omitempty"`
	WinsA  int  `json:"wins_a"`
	WinsB  int  `json:"wins_b"`
	// Coin says two bots did not play the match out: the result was a
	// coin toss.
	Coin bool `json:"coin,omitempty"`
}

func matchView(p *Pairing) MatchView {
	m := MatchView{A: p.A, Bye: p.Bye(), Done: p.Done, WinsA: p.WinsA, WinsB: p.WinsB, Coin: p.Coin}
	if !p.Bye() {
		b := p.B
		m.B = &b
	}
	if p.Done && p.Winner != NoSeat {
		w := p.Winner
		m.Winner = &w
	}
	return m
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
	// Built says the seat has a deck for the rounds. A bot always has.
	Built bool `json:"built,omitempty"`
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
		Cube:         d.cfg.Cube,
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
		Picks:        withRatings(append([]Card{}, me.picks...)),
		Phase:        d.Phase(),
		BestOf:       d.BestOf(),
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
			Built:     d.play != nil && d.play.built[i],
		})
	}
	switch d.state {
	case StatePicking:
		v.Passing = DirectionOf(d.round)
		deadline, at := d.deadline, now
		v.Deadline, v.Now = &deadline, &at
		if !me.picked {
			v.Pack = withRatings(append([]Card(nil), me.pack...))
			v.Take = d.take(len(me.pack))
		}
	case StateDone:
		v.Pool = withRatings(append([]Card{}, me.picks...))
		d.playView(&v, you, now)
	}
	return v, true
}

// Preview is what someone holding the invite sees before taking a seat.
type Preview struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Set     string    `json:"set"`
	SetName string    `json:"set_name"`
	Cube    bool      `json:"cube,omitempty"`
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
		Cube: d.cfg.Cube, Mode: d.cfg.Mode, Seats: d.cfg.Seats, State: d.state,
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

// playView fills in what follows the picks for one seat.
func (d *Draft) playView(v *View, you int, now time.Time) {
	p := d.play
	if p == nil {
		return
	}
	if p.built[you] {
		own := p.decks[you].clone()
		v.Deck = &own
	}
	if p.phase == PhaseBuilding {
		b := &BuildView{
			Now:       now,
			DeckMin:   limitedDeckMin,
			Submitted: p.built[you] && !p.auto[you],
		}
		if d.cfg.BuildUntimed {
			b.Untimed = true
		} else {
			deadline := p.buildDeadline
			b.Deadline = &deadline
			b.Seconds = int(d.buildTimeout() / time.Second)
		}
		v.Build = b
		return
	}
	if len(p.rounds) == 0 {
		return
	}
	pv := &PlayView{
		Round:     len(p.rounds),
		Rounds:    MatchRounds,
		Matches:   make([][]MatchView, 0, len(p.rounds)),
		Standings: d.Standings(),
	}
	for _, round := range p.rounds {
		row := make([]MatchView, 0, len(round))
		for _, m := range round {
			row = append(row, matchView(m))
		}
		pv.Matches = append(pv.Matches, row)
	}
	if p.phase == PhaseRounds {
		for _, m := range p.rounds[len(p.rounds)-1] {
			if !m.has(you) {
				continue
			}
			mine := matchView(m)
			pv.Yours = &mine
			if m.Table != uuid.Nil && !m.Done {
				table := m.Table
				pv.Table = &table
			}
		}
	}
	v.Play = pv
}

// limitedDeckMin is the smallest limited deck (CR 100.2b), read from
// the format's row.
var limitedDeckMin = game.FormatLimited.Spec().Deck.MinCards
