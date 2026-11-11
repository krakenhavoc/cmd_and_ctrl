// Package draft is a booster draft: seats, packs, picks and passing
// (ADR 0222 §5, #5438).
//
// A draft is its own lobby object, not a game. It never touches the
// rules engine: it hands cards round a table and, when the last pick is
// made, holds each seat's card pool. Deckbuilding and the matches the
// pools are played in are slice 6 (#5440).
//
// # The rules it follows
//
// The Comprehensive Rules describe a draft only inside the Conspiracy
// Draft variant, and this is that procedure:
//
//	CR 905.1a: "A draft typically consists of three draft rounds. In
//	each draft round, each player opens a booster pack, drafts one card
//	by placing that card in a face-down pile in front of the player,
//	then passes the remaining cards to the next player. Each player then
//	drafts a card from the booster pack passed to them and passes the
//	remaining cards. This procedure continues until all cards in that
//	draft round have been drafted."
//
//	CR 905.1b: "In the first and third draft rounds, booster packs are
//	passed to each player's left. In the second draft round, booster
//	packs are passed to each player's right."
//
//	CR 905.1c: "During the draft, a player can look only at cards in
//	the booster pack they are currently drafting from, cards they have
//	already drafted, cards that are currently revealed as described in
//	rule 905.2b, and cards that have been drafted face up as described
//	in rule 905.2c. A player may not reveal drafted cards to other
//	players unless an ability instructs them to."
//
//	CR 905.1d: "After the draft and all actions that may be taken
//	during or after the draft, all the cards a player has drafted become
//	that player's card pool. The player builds their deck from only
//	these cards and any number of basic land cards."
//
// Pick two is ADR 0222 owner decision 4, not a rule of the game: each
// pick takes two cards from the pack instead of one.
//
// Every seat picks from the pack it holds, and the packs pass together
// once every seat has picked. The player to a seat's left is the next
// higher seat number, wrapping round.
//
// # What lives where
//
// Draft is the state and its rules. It takes the time as an argument,
// holds no lock and starts no goroutine, so a test replays a whole draft
// from a seed. Room (room.go) is a Draft with a lock, a pick timer and
// the sockets watching it. View (view.go) is what one seat may see.
package draft

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Mode is how many cards a pick takes.
type Mode string

// The pick modes (ADR 0222 owner decision 4).
const (
	// PickOne: take one card, pass the rest.
	PickOne Mode = "pick_one"
	// PickTwo: take two cards, pass the rest. Offered by default at
	// DefaultPickTwoSeats seats and allowed at any size.
	PickTwo Mode = "pick_two"
)

// PerPick is how many cards one pick takes in this mode.
func (m Mode) PerPick() int {
	if m == PickTwo {
		return 2
	}
	return 1
}

// Valid reports whether m is a mode this package knows.
func (m Mode) Valid() bool { return m == PickOne || m == PickTwo }

// The bounds of a draft.
const (
	// Rounds is how many packs each seat opens (CR 905.1a).
	Rounds = 3
	// MinSeats and MaxSeats bound a draft's size (ADR 0222 §5).
	MinSeats = 2
	MaxSeats = 8
	// DefaultSeats is a full draft.
	DefaultSeats = 8
	// DefaultPickTwoSeats is the size pick two was designed for and is
	// offered at by default.
	DefaultPickTwoSeats = 4
	// DefaultPickTimeout, MinPickTimeout and MaxPickTimeout bound the
	// pick timer.
	DefaultPickTimeout = 60 * time.Second
	MinPickTimeout     = 15 * time.Second
	MaxPickTimeout     = 5 * time.Minute
)

// State is where a draft is.
type State string

// The states, in order. A draft only moves forward.
const (
	// StateLobby: seats are being taken.
	StateLobby State = "lobby"
	// StatePicking: packs are open.
	StatePicking State = "picking"
	// StateDone: the last pick is made and every seat holds its pool.
	// What follows, the build and the rounds, is Phase (limited.go).
	StateDone State = "done"
)

// Direction is which way packs pass in a round.
type Direction string

// The directions of CR 905.1b.
const (
	Left  Direction = "left"
	Right Direction = "right"
)

// DirectionOf is the direction packs pass in a round, counted from 1:
// left in rounds one and three, right in round two (CR 905.1b).
func DirectionOf(round int) Direction {
	if round%2 == 0 {
		return Right
	}
	return Left
}

// Errors a draft refuses with.
var (
	ErrStarted     = errors.New("draft: the draft has already started")
	ErrNotPicking  = errors.New("draft: no pick is being made")
	ErrFull        = errors.New("draft: every seat is taken")
	ErrNoSuchSeat  = errors.New("draft: no such seat")
	ErrNotHost     = errors.New("draft: only the host may do that")
	ErrAlreadyPick = errors.New("draft: this seat has already picked from its pack")
	ErrBadPick     = errors.New("draft: that is not a pick from the pack this seat holds")
	ErrBadConfig   = errors.New("draft: bad settings")
)

// Card is one card of a draft. Two copies of a card are two Cards with
// two IDs.
type Card struct {
	// ID names this card within the draft ("c41"). It is what a pick
	// names.
	ID string `json:"id"`
	// ScryfallID is the printing, what the image routes take.
	ScryfallID uuid.UUID `json:"scryfall_id"`
	OracleID   uuid.UUID `json:"oracle_id"`
	Name       string    `json:"name"`
	// Set and CollectorNo name the printing ("blb", "214").
	Set         string `json:"set"`
	CollectorNo string `json:"collector_number"`
	Rarity      string `json:"rarity"`
	// Colors is the card's colours as Scryfall spells them (W, U, B, R,
	// G), empty for a colourless card.
	Colors []string `json:"colors"`
	// ColorIdentity is the card's colour identity, W U B R G, from the
	// printing. It is what colours a land, whose own colours are none
	// (#5704). Absent for a colourless identity.
	ColorIdentity []string `json:"color_identity,omitempty"`
	ManaCost      string   `json:"mana_cost"`
	// ManaValue is computed from ManaCost (ManaValue).
	ManaValue int    `json:"mana_value"`
	TypeLine  string `json:"type_line"`
	// Basic marks a basic land.
	Basic bool `json:"basic,omitempty"`
}

// Source makes the packs of a draft. Each call returns one new pack; the
// same rng state gives the same pack. The cards' IDs are assigned by the
// draft, so a Source leaves them empty.
type Source interface {
	Pack(rng *rand.Rand) ([]Card, error)
}

// Dealer is a Source whose packs depend on each other, so a draft takes
// all of them in one call: a cube, whose packs are cut from one shuffle
// of its list and share no card (CubeSource). Start asks a Dealer for
// n packs and hands them out in the order it would have drawn them,
// round by round and seat by seat.
type Dealer interface {
	Source
	Deal(rng *rand.Rand, n int) ([][]Card, error)
}

// Seat is one seat of a draft.
type Seat struct {
	// ID is the seat's identity for the life of the draft. It is never
	// sent to another seat; a view names seats by their number.
	ID   uuid.UUID
	Name string
	// Bot is true for a seat no person holds: one the start filled, or
	// one a person left during the picks.
	Bot bool
	// UserID is the signed-in person in the seat, uuid.Nil for a guest
	// or a bot. It is what lets that person take the seat back from
	// another device. Held in memory only.
	UserID uuid.UUID

	// pack is the pack the seat holds and has not yet picked from. Nil
	// once it has picked and is waiting for the others.
	pack []Card
	// rest is what is left of the pack after this seat's pick, waiting
	// to be passed.
	rest []Card
	// picked is true once the seat has picked in the current step.
	picked bool
	// picks is every card the seat has drafted, in the order taken.
	picks []Card
}

// Config is what a draft is made with.
type Config struct {
	Name string
	// Set is the lowercase set code, SetName its name and Booster the
	// kind of booster its packs approximate. They are labels here; the
	// packs come from the Source given to Start.
	Set     string
	SetName string
	Booster string
	// Cube marks a draft of a saved card list (ADR 0222 §8): Set is
	// empty, SetName is the cube's name and Booster is "cube".
	Cube bool
	Mode Mode
	// Seats is the size of the draft, MinSeats to MaxSeats.
	Seats int
	// Seed is what the packs are drawn from, and what the bot's picks are
	// a function of. Two drafts with one seed, one Source and the same
	// picks are the same draft.
	Seed uint64
	// PickTimeout is the pick timer. Zero is DefaultPickTimeout.
	PickTimeout time.Duration
	// Solo marks a draft made by a visitor: no invite, and every other
	// seat is a bot.
	Solo bool
	// Picker picks for bot seats, for a seat whose timer ran out and for
	// a seat that is away. Nil is BotPicker{Seed}, the draft bot (#5439).
	Picker Picker

	// What follows the picks (limited.go, #5440).
	//
	// BuildTimeout is the deckbuilding timer. Zero is
	// DefaultBuildTimeout.
	BuildTimeout time.Duration
	// BestOf is the length of the draft's matches, 1 or 3. Zero is 1.
	BestOf int
	// Builder builds a deck for a bot seat and for a person whose build
	// timer ran out. Nil is BotDeckBuilder{Seed}, the draft bot (#5439).
	Builder DeckBuilder
}

// Draft is one booster draft. It is not safe for concurrent use; Room
// wraps it.
type Draft struct {
	ID        uuid.UUID
	CreatedAt time.Time
	cfg       Config
	picker    Picker

	state State
	seats []*Seat
	host  uuid.UUID

	// unopened[round][seat] is the pack that seat opens in that round
	// (both from 0). Filled by Start.
	unopened [][][]Card
	// round is the current round from 1, pick the pick within it from 1.
	round, pick int
	// picksInRound is how many picks the current round has.
	picksInRound int
	// deadline is when the current pick's timer runs out.
	deadline time.Time
	// stepAt is when the current pick began.
	stepAt time.Time
	// seq rises with every change a view can show.
	seq uint64

	// play is the build and the rounds (limited.go), nil until the last
	// pick is made.
	play *play
}

// New makes a draft in its lobby, with no seat taken.
func New(cfg Config, now time.Time) (*Draft, error) {
	if !cfg.Mode.Valid() {
		return nil, fmt.Errorf("%w: unknown pick mode %q", ErrBadConfig, cfg.Mode)
	}
	if cfg.Seats < MinSeats || cfg.Seats > MaxSeats {
		return nil, fmt.Errorf("%w: %d seats (want %d to %d)", ErrBadConfig, cfg.Seats, MinSeats, MaxSeats)
	}
	if cfg.PickTimeout == 0 {
		cfg.PickTimeout = DefaultPickTimeout
	}
	if cfg.PickTimeout < 0 {
		return nil, fmt.Errorf("%w: a negative pick timer", ErrBadConfig)
	}
	if cfg.BuildTimeout < 0 {
		return nil, fmt.Errorf("%w: a negative build timer", ErrBadConfig)
	}
	if cfg.BestOf != 0 && cfg.BestOf != 1 && cfg.BestOf != 3 {
		return nil, fmt.Errorf("%w: best of %d (want 1 or 3)", ErrBadConfig, cfg.BestOf)
	}
	d := &Draft{ID: uuid.New(), CreatedAt: now, cfg: cfg, state: StateLobby, picker: cfg.Picker}
	if d.picker == nil {
		d.picker = BotPicker{Seed: cfg.Seed}
	}
	return d, nil
}

// Config returns the settings the draft was made with.
func (d *Draft) Config() Config { return d.cfg }

// State is where the draft is.
func (d *Draft) State() State { return d.state }

// Seed is the seed the draft's packs are drawn from.
func (d *Draft) Seed() uint64 { return d.cfg.Seed }

// Round and PickNumber are the current round and the pick within it,
// both from 1, and 0 outside the picks.
func (d *Draft) Round() int      { return d.round }
func (d *Draft) PickNumber() int { return d.pick }

// Deadline is when the current pick's timer runs out, zero outside the
// picks.
func (d *Draft) Deadline() time.Time { return d.deadline }

// StepAt is when the current pick began, zero outside the picks.
func (d *Draft) StepAt() time.Time { return d.stepAt }

// Seq rises with every change.
func (d *Draft) Seq() uint64 { return d.seq }

// Seats is the seats in order. The slice is a copy; the seats are the
// draft's own.
func (d *Draft) Seats() []*Seat { return append([]*Seat(nil), d.seats...) }

// Host is the seat that may start the draft, uuid.Nil when no person is
// seated.
func (d *Draft) Host() uuid.UUID { return d.host }

// Humans counts the seats a person holds.
func (d *Draft) Humans() int {
	n := 0
	for _, s := range d.seats {
		if !s.Bot {
			n++
		}
	}
	return n
}

func (d *Draft) seat(id uuid.UUID) (int, *Seat) {
	for i, s := range d.seats {
		if s.ID == id {
			return i, s
		}
	}
	return -1, nil
}

// SeatNumber is a seat's number from 0, or -1.
func (d *Draft) SeatNumber(id uuid.UUID) int {
	i, _ := d.seat(id)
	return i
}

// SeatOfUser is the seat a signed-in person holds, or uuid.Nil.
func (d *Draft) SeatOfUser(userID uuid.UUID) uuid.UUID {
	if userID == uuid.Nil {
		return uuid.Nil
	}
	for _, s := range d.seats {
		if !s.Bot && s.UserID == userID {
			return s.ID
		}
	}
	return uuid.Nil
}

// AddSeat seats a person while the draft is in its lobby. The first
// person seated is the host.
func (d *Draft) AddSeat(name string, userID uuid.UUID) (*Seat, error) {
	if d.state != StateLobby {
		return nil, ErrStarted
	}
	if len(d.seats) >= d.cfg.Seats {
		return nil, ErrFull
	}
	s := &Seat{ID: uuid.New(), Name: name, UserID: userID}
	d.seats = append(d.seats, s)
	if d.host == uuid.Nil {
		d.host = s.ID
	}
	d.seq++
	return s, nil
}

// Leave takes a person out of their seat. In the lobby the seat is
// freed and, if they were the host, the next person hosts. During the
// picks the seat becomes a bot, which picks for it from then on, and
// its pick for the pack it holds is made at once. A finished draft keeps
// the seat's pool and only forgets the person.
func (d *Draft) Leave(seatID uuid.UUID, now time.Time) error {
	i, s := d.seat(seatID)
	if s == nil || s.Bot {
		return ErrNoSuchSeat
	}
	switch d.state {
	case StateLobby:
		d.seats = append(d.seats[:i], d.seats[i+1:]...)
	default:
		s.Bot = true
		s.UserID = uuid.Nil
	}
	if d.host == seatID {
		d.host = uuid.Nil
		for _, o := range d.seats {
			if !o.Bot {
				d.host = o.ID
				break
			}
		}
	}
	d.seq++
	switch d.state {
	case StatePicking:
		d.advance(now)
	case StateDone:
		d.afterLeave(i)
	}
	return nil
}

// Start fills every seat no person has taken with a bot, draws every
// pack from the seed and opens the first round. by is the seat asking,
// which must be the host.
func (d *Draft) Start(by uuid.UUID, src Source, now time.Time) error {
	if d.state != StateLobby {
		return ErrStarted
	}
	if by != d.host || by == uuid.Nil {
		return ErrNotHost
	}
	people := len(d.seats)
	for n := 1; len(d.seats) < d.cfg.Seats; n++ {
		d.seats = append(d.seats, &Seat{ID: uuid.New(), Name: "Bot " + strconv.Itoa(n), Bot: true})
	}
	packs, err := d.dealPacks(src)
	if err != nil {
		// A refused start leaves the lobby as it was.
		d.seats = d.seats[:people]
		return err
	}
	d.unopened = packs
	d.state = StatePicking
	d.round = 0
	d.openRound(now)
	d.advance(now)
	return nil
}

// dealPacks draws every pack of the draft: packs[round][seat].
func (d *Draft) dealPacks(src Source) ([][][]Card, error) {
	// One stream for every pack, drawn in a fixed order (round, then
	// seat), so the seed alone decides who opens what.
	rng := rand.New(rand.NewPCG(d.cfg.Seed, packStream))
	seats := len(d.seats)
	var dealt [][]Card
	if dealer, ok := src.(Dealer); ok {
		var err error
		if dealt, err = dealer.Deal(rng, Rounds*seats); err != nil {
			return nil, err
		}
		if len(dealt) != Rounds*seats {
			return nil, fmt.Errorf("draft: the pack source dealt %d packs, want %d", len(dealt), Rounds*seats)
		}
	}
	next := 1
	packs := make([][][]Card, Rounds)
	for r := range Rounds {
		packs[r] = make([][]Card, seats)
		for s := range seats {
			var pack []Card
			if dealt != nil {
				pack = dealt[r*seats+s]
			} else {
				var err error
				if pack, err = src.Pack(rng); err != nil {
					return nil, err
				}
			}
			if len(pack) == 0 {
				return nil, fmt.Errorf("draft: the pack source made an empty pack")
			}
			for i := range pack {
				pack[i].ID = "c" + strconv.Itoa(next)
				next++
			}
			packs[r][s] = pack
		}
	}
	return packs, nil
}

// packStream is the PCG stream the packs are drawn on. The placeholder
// picker draws on others (RandomPicker), so a change to how bots pick
// never changes the packs a seed gives.
const packStream = 0x5438

// openRound hands every seat its next unopened pack.
func (d *Draft) openRound(now time.Time) {
	d.round++
	d.pick = 1
	most := 0
	for i, s := range d.seats {
		s.pack = d.unopened[d.round-1][i]
		s.rest = nil
		s.picked = false
		most = max(most, len(s.pack))
	}
	d.unopened[d.round-1] = nil
	per := d.cfg.Mode.PerPick()
	d.picksInRound = (most + per - 1) / per
	d.beginStep(now)
}

func (d *Draft) beginStep(now time.Time) {
	d.stepAt = now
	d.deadline = now.Add(d.cfg.PickTimeout)
	d.seq++
}

// take is how many cards a seat must take from a pack of n cards.
func (d *Draft) take(n int) int { return min(d.cfg.Mode.PerPick(), n) }

// Take is how many cards the seat must take from the pack it holds now:
// 0 when it holds none.
func (d *Draft) Take(seatID uuid.UUID) int {
	_, s := d.seat(seatID)
	if s == nil || d.state != StatePicking || s.picked {
		return 0
	}
	return d.take(len(s.pack))
}

// Pending is the seats that have not yet picked from the pack they
// hold, in seat order. Empty outside the picks.
func (d *Draft) Pending() []*Seat {
	if d.state != StatePicking {
		return nil
	}
	var out []*Seat
	for _, s := range d.seats {
		if !s.picked {
			out = append(out, s)
		}
	}
	return out
}

// Pick takes cards for a seat from the pack it holds: exactly Take
// different cards of that pack, named by ID. When the last seat has
// picked, the packs pass.
func (d *Draft) Pick(seatID uuid.UUID, ids []string, now time.Time) error {
	if d.state != StatePicking {
		return ErrNotPicking
	}
	_, s := d.seat(seatID)
	if s == nil {
		return ErrNoSuchSeat
	}
	if err := d.apply(s, ids); err != nil {
		return err
	}
	d.advance(now)
	return nil
}

// apply moves the named cards from a seat's pack to its picks.
func (d *Draft) apply(s *Seat, ids []string) error {
	if s.picked {
		return ErrAlreadyPick
	}
	want := d.take(len(s.pack))
	if len(ids) != want {
		return fmt.Errorf("%w: %d cards named, this pick takes %d", ErrBadPick, len(ids), want)
	}
	chosen := map[string]bool{}
	for _, id := range ids {
		if chosen[id] {
			return fmt.Errorf("%w: %s is named twice", ErrBadPick, id)
		}
		chosen[id] = true
	}
	var taken, rest []Card
	for _, c := range s.pack {
		if chosen[c.ID] {
			taken = append(taken, c)
		} else {
			rest = append(rest, c)
		}
	}
	if len(taken) != want {
		return fmt.Errorf("%w: a named card is not in the pack", ErrBadPick)
	}
	// In the order the seat named them.
	for _, id := range ids {
		for _, c := range taken {
			if c.ID == id {
				s.picks = append(s.picks, c)
			}
		}
	}
	s.pack, s.rest, s.picked = nil, rest, true
	d.seq++
	return nil
}

// AutoPick makes the Picker's pick for a seat that has not picked: a
// seat whose timer ran out, or one that is away. It does nothing for a
// seat that has picked.
func (d *Draft) AutoPick(seatID uuid.UUID, now time.Time) {
	if d.state != StatePicking {
		return
	}
	i, s := d.seat(seatID)
	if s == nil || s.picked {
		return
	}
	d.pickFor(i, s)
	d.advance(now)
}

// Expire makes the Picker's pick for every seat that has not picked. It
// is what the pick timer does when it runs out.
func (d *Draft) Expire(now time.Time) {
	if d.state != StatePicking {
		return
	}
	for i, s := range d.seats {
		if !s.picked {
			d.pickFor(i, s)
		}
	}
	d.advance(now)
}

// pickFor asks the Picker and applies its answer. A Picker that answers
// with something that is not a legal pick is overruled with the first
// cards of the pack: a draft never stalls on a bot.
func (d *Draft) pickFor(i int, s *Seat) {
	want := d.take(len(s.pack))
	if want == 0 {
		s.pack, s.rest, s.picked = nil, nil, true
		return
	}
	ids := d.picker.Pick(PickRequest{
		Pack:   append([]Card(nil), s.pack...),
		Picked: append([]Card(nil), s.picks...),
		Mode:   d.cfg.Mode,
		Take:   want,
		Seat:   i,
		Seats:  len(d.seats),
		Round:  d.round,
		Pick:   d.pick,
	})
	if d.apply(s, ids) == nil {
		return
	}
	first := make([]string, 0, want)
	for _, c := range s.pack[:want] {
		first = append(first, c.ID)
	}
	_ = d.apply(s, first)
}

// advance runs the draft forward as far as it goes without a person:
// bots pick, and once every seat has picked the packs pass, the next
// round opens or the draft is done.
func (d *Draft) advance(now time.Time) {
	for d.state == StatePicking {
		for i, s := range d.seats {
			if s.picked {
				continue
			}
			if s.Bot || len(s.pack) == 0 {
				d.pickFor(i, s)
			}
		}
		for _, s := range d.seats {
			if !s.picked {
				return
			}
		}
		d.pass(now)
	}
}

// pass hands every seat the rest of its neighbour's pack: from the
// right-hand neighbour when packs pass left, from the left-hand one
// when they pass right (CR 905.1b). With nothing left to pass, the
// round is over.
func (d *Draft) pass(now time.Time) {
	n := len(d.seats)
	left := 0
	for _, s := range d.seats {
		left += len(s.rest)
	}
	if left == 0 {
		if d.round >= Rounds {
			d.state = StateDone
			d.pick, d.picksInRound = 0, 0
			d.deadline, d.stepAt = time.Time{}, time.Time{}
			for _, s := range d.seats {
				s.pack, s.rest, s.picked = nil, nil, false
			}
			d.seq++
			d.beginBuilding(now)
			return
		}
		d.openRound(now)
		return
	}
	step := 1
	if DirectionOf(d.round) == Right {
		step = -1
	}
	rests := make([][]Card, n)
	for i, s := range d.seats {
		rests[i] = s.rest
	}
	for i, s := range d.seats {
		// Seat i passes to seat i+step, so it receives from i-step.
		s.pack = rests[((i-step)%n+n)%n]
		s.rest = nil
		s.picked = false
	}
	d.pick++
	d.beginStep(now)
}

// Pool is every card a seat has drafted, in the order taken: its card
// pool once the draft is done (CR 905.1d), and its picks so far before
// that. This is what slice 6 builds a deck from.
func (d *Draft) Pool(seatID uuid.UUID) []Card {
	_, s := d.seat(seatID)
	if s == nil {
		return nil
	}
	return append([]Card(nil), s.picks...)
}

// Pools is every seat's pool, by seat number. It is the server's own
// reading (deckbuilding, bot decks, tests); no view carries it.
func (d *Draft) Pools() [][]Card {
	out := make([][]Card, len(d.seats))
	for i, s := range d.seats {
		out[i] = append([]Card(nil), s.picks...)
	}
	return out
}

// HeldPack is the pack a seat holds and has not picked from, nil when
// it holds none. The server's own reading, like Pools.
func (d *Draft) HeldPack(seatID uuid.UUID) []Card {
	_, s := d.seat(seatID)
	if s == nil {
		return nil
	}
	return append([]Card(nil), s.pack...)
}
