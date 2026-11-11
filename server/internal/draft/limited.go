package draft

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// limited.go is what a draft does once its last pick is made (ADR 0222
// §5 and §6, #5440): each seat builds a deck from its pool, and the
// decks play rounds of one-against-one matches.
//
// The draft's State stays StateDone through all of it, which is what it
// says: the picks are done and every seat holds its pool. Phase says
// where the draft is after that.
//
// # Building
//
// Every seat builds a deck of at least 40 cards from its pool and any
// number of basic lands (CR 100.2b, 905.1d; the check is
// deck.ValidateLimited). A bot's deck is built at once by the
// DeckBuilder. A person submits theirs, and may submit again until the
// rounds begin, which is when the last person has submitted or the
// build timer runs out. A person who has not submitted by then gets
// the DeckBuilder's deck from their pool.
//
// # Rounds
//
// MatchRounds rounds. Each round pairs every seat with another, at
// random from the draft's seed, by these preferences in order: no seat
// meets a seat it has already played; people are paired with people;
// with an odd number of seats the seat that sits the round out is a
// bot, and one that has not sat out before. A draft of two seats has
// one possible pairing, so its later rounds are rematches: there is
// nobody else to play.
//
// A match with a person in it is played at a limited table the lobby
// opens; the draft only records its result. A match between two bots
// is not played: its result is a coin tossed from the seed (ADR 0222
// §9 and Calls made here), decided as the round begins so no person
// waits on it. The next round begins when every match of this one has
// a result.
//
// A person may leave after any round, or during one. Their seat
// becomes a bot, as it does during the picks, and plays on with the
// deck they built. The draft is finished when its rounds are, or when
// no person is left in it.

// Phase is where a draft is after its last pick.
type Phase string

// The phases, in order.
const (
	// PhaseBuilding: decks are being built.
	PhaseBuilding Phase = "building"
	// PhaseRounds: matches are being played.
	PhaseRounds Phase = "rounds"
	// PhaseFinished: the rounds are over, or every person has left.
	PhaseFinished Phase = "finished"
)

// The bounds of what follows the picks.
const (
	// MatchRounds is how many rounds of matches a draft plays (ADR 0222
	// §6).
	MatchRounds = 3
	// DefaultBuildTimeout, MinBuildTimeout and MaxBuildTimeout bound
	// the build timer.
	DefaultBuildTimeout = 15 * time.Minute
	MinBuildTimeout     = 2 * time.Minute
	MaxBuildTimeout     = 30 * time.Minute
)

// Errors of the build and the rounds.
var (
	ErrNotBuilding = errors.New("draft: decks are not being built")
	ErrNoSuchMatch = errors.New("draft: no such match")
	// ErrDeckIllegal wraps the reason a submitted deck is refused. The
	// seat's last legal deck, if it had one, stands.
	ErrDeckIllegal = errors.New("draft: that deck is not legal")
)

// NoSeat is the seat number of nobody: a bye's missing opponent, and
// the winner of a drawn match.
const NoSeat = -1

// Pairing is one match of a round.
type Pairing struct {
	// A and B are the seats, by number. B is NoSeat for a bye: with an
	// odd number of seats, A sits the round out.
	A, B int
	// BotA and BotB say whether each seat was a bot when the round was
	// paired, and UserA and UserB the signed-in person in it then
	// (uuid.Nil for a guest or a bot). They are what a stored result
	// names its opponent by.
	BotA, BotB   bool
	UserA, UserB uuid.UUID
	// Table is the limited table the match is played at, uuid.Nil until
	// the lobby opens it and for a match that is not played at one.
	Table uuid.UUID
	// Done says the match has a result. Winner is the seat that won it,
	// NoSeat for a draw or a bye. WinsA and WinsB are the games each
	// seat won.
	Done         bool
	Winner       int
	WinsA, WinsB int
	// Coin says the result was a coin toss between two bots, not a
	// match that was played.
	Coin bool
}

// Bye reports whether the pairing is a seat sitting the round out.
func (p *Pairing) Bye() bool { return p.B == NoSeat }

// has reports whether seat plays in the pairing.
func (p *Pairing) has(seat int) bool { return p.A == seat || p.B == seat }

// play is the state of a draft after its picks.
type play struct {
	phase         Phase
	buildDeadline time.Time
	buildAt       time.Time
	// decks[seat] is the seat's deck, valid once built[seat]. auto says
	// the DeckBuilder built it.
	decks []LimitedDeck
	built []bool
	auto  []bool
	// rounds is every round paired so far; the last is the one in play.
	rounds [][]*Pairing
	// person[seat] says a person held the seat when the picks ended.
	person []bool
}

// Phase is where the draft is after its picks, "" before they are done.
func (d *Draft) Phase() Phase {
	if d.play == nil {
		return ""
	}
	return d.play.phase
}

// BuildDeadline is when the build timer runs out, zero outside the
// build.
func (d *Draft) BuildDeadline() time.Time {
	if d.Phase() != PhaseBuilding {
		return time.Time{}
	}
	return d.play.buildDeadline
}

// BestOf is the length of the draft's matches: 1 or 3.
func (d *Draft) BestOf() int {
	if d.cfg.BestOf == 3 {
		return 3
	}
	return 1
}

func (d *Draft) builder() DeckBuilder {
	if d.cfg.Builder != nil {
		return d.cfg.Builder
	}
	return SimpleDeckBuilder{}
}

func (d *Draft) buildTimeout() time.Duration {
	if d.cfg.BuildTimeout > 0 {
		return d.cfg.BuildTimeout
	}
	return DefaultBuildTimeout
}

// beginBuilding opens the build as the last pick is made. Every bot's
// deck is built at once; a draft with no person in it has nothing to
// wait for and is finished.
func (d *Draft) beginBuilding(now time.Time) {
	n := len(d.seats)
	p := &play{
		phase:         PhaseBuilding,
		buildAt:       now,
		buildDeadline: now.Add(d.buildTimeout()),
		decks:         make([]LimitedDeck, n),
		built:         make([]bool, n),
		auto:          make([]bool, n),
		person:        make([]bool, n),
	}
	d.play = p
	for i, s := range d.seats {
		p.person[i] = !s.Bot
		if s.Bot {
			d.autoBuild(i)
		}
	}
	d.seq++
	d.afterBuild()
}

// autoBuild gives a seat the DeckBuilder's deck. A builder whose deck
// fails the limited check is overruled by the plain one, and a pool
// even that cannot make a deck from (it cannot happen: basics are
// free) is topped up with basic lands.
func (d *Draft) autoBuild(seat int) {
	pool := d.seats[seat].picks
	built := d.builder().BuildLimitedDeck(append([]Card(nil), pool...))
	if d.checkDeck(pool, built) != nil {
		built = SimpleDeckBuilder{}.BuildLimitedDeck(append([]Card(nil), pool...))
	}
	if d.checkDeck(pool, built) != nil {
		built = LimitedDeck{Basics: map[string]int{BasicLandNames[0]: 40}}
	}
	d.play.decks[seat] = built.clone()
	d.play.built[seat] = true
	d.play.auto[seat] = true
}

// checkDeck is the limited deck check (deck.ValidateLimited) on a deck
// named by draft IDs. A card named twice, or one the pool does not
// hold, is refused before the rule is asked: an ID is one card.
func (d *Draft) checkDeck(pool []Card, built LimitedDeck) error {
	byID := make(map[string]Card, len(pool))
	for _, c := range pool {
		byID[c.ID] = c
	}
	seen := map[string]bool{}
	main := make([]deck.LimitedCard, 0, built.Size())
	for _, id := range built.Main {
		c, ok := byID[id]
		if !ok {
			return fmt.Errorf("%w: a card of the deck is not in your card pool", ErrDeckIllegal)
		}
		if seen[id] {
			return fmt.Errorf("%w: %q is in the deck twice, and your card pool holds it once", ErrDeckIllegal, c.Name)
		}
		seen[id] = true
		main = append(main, deck.LimitedCard{Name: c.Name, Basic: c.Basic})
	}
	for name, n := range built.Basics {
		if !IsBasicLandName(name) {
			return fmt.Errorf("%w: %q is not a basic land", ErrDeckIllegal, name)
		}
		if n < 0 || n > maxBasicsOfOne {
			return fmt.Errorf("%w: %d %s (want 0 to %d)", ErrDeckIllegal, n, name, maxBasicsOfOne)
		}
		for range n {
			main = append(main, deck.LimitedCard{Name: name, Basic: true})
		}
	}
	all := make([]deck.LimitedCard, 0, len(pool))
	for _, c := range pool {
		all = append(all, deck.LimitedCard{Name: c.Name, Basic: c.Basic})
	}
	if err := deck.ValidateLimited(main, all); err != nil {
		var ve *deck.ValidationError
		if errors.As(err, &ve) && len(ve.Violations) > 0 {
			return fmt.Errorf("%w: %s", ErrDeckIllegal, ve.Violations[0].Message)
		}
		return fmt.Errorf("%w: %v", ErrDeckIllegal, err)
	}
	return nil
}

// maxBasicsOfOne bounds a request, not the rule: the rule allows any
// number of basic lands, and no deck wants a hundred of one.
const maxBasicsOfOne = 99

// SubmitDeck is a person's deck for the rounds. It is checked against
// the seat's pool; a deck that fails is refused and changes nothing.
// A person may submit again until the rounds begin, which is when the
// last person has submitted.
func (d *Draft) SubmitDeck(seatID uuid.UUID, built LimitedDeck, _ time.Time) error {
	if d.Phase() != PhaseBuilding {
		return ErrNotBuilding
	}
	i, s := d.seat(seatID)
	if s == nil || s.Bot {
		return ErrNoSuchSeat
	}
	if err := d.checkDeck(s.picks, built); err != nil {
		return err
	}
	d.play.decks[i] = built.clone()
	d.play.built[i] = true
	d.play.auto[i] = false
	d.seq++
	d.afterBuild()
	return nil
}

// ExpireBuild ends the build when its timer has run out: every person
// who has not submitted gets the DeckBuilder's deck, and the rounds
// begin. It does nothing before the deadline.
func (d *Draft) ExpireBuild(now time.Time) {
	if d.Phase() != PhaseBuilding || now.Before(d.play.buildDeadline) {
		return
	}
	for i := range d.seats {
		if !d.play.built[i] {
			d.autoBuild(i)
		}
	}
	d.seq++
	d.afterBuild()
}

// afterBuild begins the rounds once every seat has a deck.
func (d *Draft) afterBuild() {
	if d.Phase() != PhaseBuilding {
		return
	}
	if d.Humans() == 0 {
		d.finish()
		return
	}
	for _, ok := range d.play.built {
		if !ok {
			return
		}
	}
	d.play.phase = PhaseRounds
	d.play.buildDeadline = time.Time{}
	d.seq++
	d.nextRound()
}

func (d *Draft) finish() {
	if d.play == nil || d.play.phase == PhaseFinished {
		return
	}
	d.play.phase = PhaseFinished
	d.play.buildDeadline = time.Time{}
	d.seq++
}

// Deck is a seat's deck by seat number, and whether it has one. The
// server's own reading, like Pools: no view carries another seat's.
func (d *Draft) Deck(seat int) (LimitedDeck, bool) {
	if d.play == nil || seat < 0 || seat >= len(d.play.built) || !d.play.built[seat] {
		return LimitedDeck{}, false
	}
	return d.play.decks[seat].clone(), true
}

// DeckCards splits a seat's pool by its deck: the pool cards it plays,
// the basic lands it adds by name, and the rest of the pool, which is
// its sideboard (CR 100.4b).
func (d *Draft) DeckCards(seat int) (main []Card, basics map[string]int, side []Card, ok bool) {
	built, ok := d.Deck(seat)
	if !ok {
		return nil, nil, nil, false
	}
	in := make(map[string]bool, len(built.Main))
	for _, id := range built.Main {
		in[id] = true
	}
	for _, c := range d.seats[seat].picks {
		if in[c.ID] {
			main = append(main, c)
		} else {
			side = append(side, c)
		}
	}
	return main, built.Basics, side, true
}

// MatchRound is the round in play from 1, 0 before the rounds.
func (d *Draft) MatchRound() int {
	if d.play == nil {
		return 0
	}
	return len(d.play.rounds)
}

// Pairings is one round's matches, round from 1. The slice is a copy.
func (d *Draft) Pairings(round int) []Pairing {
	if d.play == nil || round < 1 || round > len(d.play.rounds) {
		return nil
	}
	out := make([]Pairing, 0, len(d.play.rounds[round-1]))
	for _, p := range d.play.rounds[round-1] {
		out = append(out, *p)
	}
	return out
}

// The PCG streams of the rounds. Like the packs' and the placeholder
// picker's, each is its own, so nothing here changes the packs a seed
// gives or the picks of its bots.
const (
	pairStream = 0x5440_0000
	coinStream = 0x5440_c000
)

// nextRound pairs the next round, tosses its bot matches, and moves on
// again if that left nothing to play. After the last round the draft
// is finished.
func (d *Draft) nextRound() {
	for d.Phase() == PhaseRounds {
		if d.Humans() == 0 || len(d.play.rounds) >= MatchRounds {
			d.finish()
			return
		}
		round := len(d.play.rounds) + 1
		rng := rand.New(rand.NewPCG(d.cfg.Seed, pairStream+uint64(round)))
		pairs := d.pairRound(rng)
		d.play.rounds = append(d.play.rounds, pairs)
		d.seq++
		d.settleBots()
		if !d.roundDone() {
			return
		}
	}
}

// pairRound chooses the round's pairings: of every way to pair the
// seats, one of the best by pairScore, at random.
func (d *Draft) pairRound(rng *rand.Rand) []*Pairing {
	n := len(d.seats)
	played := map[[2]int]bool{}
	satOut := make([]bool, n)
	for _, round := range d.play.rounds {
		for _, p := range round {
			if p.Bye() {
				satOut[p.A] = true
				continue
			}
			played[pairKey(p.A, p.B)] = true
		}
	}
	bot := make([]bool, n)
	for i, s := range d.seats {
		bot[i] = s.Bot
	}

	var best [][][2]int
	bestScore := [4]int{1 << 30}
	used := make([]bool, n)
	var cur [][2]int
	var walk func(byeLeft bool)
	walk = func(byeLeft bool) {
		a := -1
		for i := range n {
			if !used[i] {
				a = i
				break
			}
		}
		if a < 0 {
			score := pairScore(cur, bot, played, satOut)
			switch cmpScore(score, bestScore) {
			case -1:
				bestScore = score
				best = [][][2]int{append([][2]int(nil), cur...)}
			case 0:
				best = append(best, append([][2]int(nil), cur...))
			}
			return
		}
		used[a] = true
		if byeLeft {
			cur = append(cur, [2]int{a, NoSeat})
			walk(false)
			cur = cur[:len(cur)-1]
		}
		for b := a + 1; b < n; b++ {
			if used[b] {
				continue
			}
			used[b] = true
			cur = append(cur, [2]int{a, b})
			walk(byeLeft)
			cur = cur[:len(cur)-1]
			used[b] = false
		}
		used[a] = false
	}
	walk(n%2 == 1)

	chosen := best[rng.IntN(len(best))]
	out := make([]*Pairing, 0, len(chosen))
	for _, c := range chosen {
		a, b := c[0], c[1]
		// Which seat is A is a coin too, so seat order means nothing.
		if b != NoSeat && rng.IntN(2) == 1 {
			a, b = b, a
		}
		p := &Pairing{A: a, B: b, Winner: NoSeat, BotA: bot[a], UserA: d.seats[a].UserID}
		if b == NoSeat {
			// A bye is a result at once: nothing is played.
			p.Done = true
		} else {
			p.BotB, p.UserB = bot[b], d.seats[b].UserID
		}
		out = append(out, p)
	}
	return out
}

func pairKey(a, b int) [2]int {
	if a > b {
		a, b = b, a
	}
	return [2]int{a, b}
}

// pairScore rates one way to pair a round; lower is better, compared
// field by field: matches already played, then (negated) matches
// between two people, so people meet people first, then people sitting
// out, then seats sitting out a second time.
func pairScore(pairs [][2]int, bot []bool, played map[[2]int]bool, satOut []bool) [4]int {
	var s [4]int
	for _, p := range pairs {
		a, b := p[0], p[1]
		if b == NoSeat {
			if !bot[a] {
				s[2]++
			}
			if satOut[a] {
				s[3]++
			}
			continue
		}
		if played[pairKey(a, b)] {
			s[0]++
		}
		if !bot[a] && !bot[b] {
			s[1]--
		}
	}
	return s
}

func cmpScore(a, b [4]int) int {
	for i := range a {
		switch {
		case a[i] < b[i]:
			return -1
		case a[i] > b[i]:
			return 1
		}
	}
	return 0
}

// settleBots tosses the coin for every match of the round in play that
// two bots are left to play and no table was opened for.
func (d *Draft) settleBots() {
	round := len(d.play.rounds)
	if round == 0 {
		return
	}
	need := d.BestOf()/2 + 1
	for i, p := range d.play.rounds[round-1] {
		if p.Done || p.Bye() || p.Table != uuid.Nil {
			continue
		}
		if !d.seats[p.A].Bot || !d.seats[p.B].Bot {
			continue
		}
		rng := rand.New(rand.NewPCG(d.cfg.Seed, coinStream+uint64(round)<<8+uint64(i)))
		p.Done, p.Coin = true, true
		if rng.IntN(2) == 0 {
			p.Winner, p.WinsA = p.A, need
		} else {
			p.Winner, p.WinsB = p.B, need
		}
		d.seq++
	}
}

func (d *Draft) roundDone() bool {
	round := len(d.play.rounds)
	if round == 0 {
		return false
	}
	for _, p := range d.play.rounds[round-1] {
		if !p.Done {
			return false
		}
	}
	return true
}

// TableNeed is a match of the round in play that the lobby must open a
// table for: it has a person in it and no table yet.
type TableNeed struct {
	Round, Index int
	A, B         int
}

// TablesNeeded lists the matches waiting for a table.
func (d *Draft) TablesNeeded() []TableNeed {
	if d.Phase() != PhaseRounds {
		return nil
	}
	round := len(d.play.rounds)
	if round == 0 {
		return nil
	}
	var out []TableNeed
	for i, p := range d.play.rounds[round-1] {
		if p.Done || p.Bye() || p.Table != uuid.Nil {
			continue
		}
		if d.seats[p.A].Bot && d.seats[p.B].Bot {
			continue
		}
		out = append(out, TableNeed{Round: round, Index: i, A: p.A, B: p.B})
	}
	return out
}

func (d *Draft) pairing(round, index int) *Pairing {
	if d.play == nil || round < 1 || round > len(d.play.rounds) {
		return nil
	}
	ps := d.play.rounds[round-1]
	if index < 0 || index >= len(ps) {
		return nil
	}
	return ps[index]
}

// SetTable records the table the lobby opened for a match.
func (d *Draft) SetTable(round, index int, table uuid.UUID) error {
	p := d.pairing(round, index)
	if p == nil || p.Done || p.Bye() {
		return ErrNoSuchMatch
	}
	p.Table = table
	d.seq++
	return nil
}

// ReportMatch records a played match's result: the seat that won it
// (NoSeat for a draw) and the games each seat won. A result for a
// match that already has one changes nothing. When it was the round's
// last, the next round begins.
func (d *Draft) ReportMatch(round, index, winner, winsA, winsB int, _ time.Time) error {
	p := d.pairing(round, index)
	if p == nil || p.Bye() || (winner != NoSeat && !p.has(winner)) {
		return ErrNoSuchMatch
	}
	if p.Done {
		return nil
	}
	p.Done, p.Winner, p.WinsA, p.WinsB = true, winner, winsA, winsB
	d.seq++
	if round == len(d.play.rounds) && d.roundDone() {
		d.nextRound()
	}
	return nil
}

// afterLeave is run when a person leaves after the picks: their seat is
// a bot now. During the build it is built for; during the rounds a
// match it had not sat down to is tossed for if a bot is all that is
// left of it. With nobody left the draft is finished.
func (d *Draft) afterLeave(seat int) {
	switch d.Phase() {
	case PhaseBuilding:
		if !d.play.built[seat] {
			d.autoBuild(seat)
		}
		d.afterBuild()
	case PhaseRounds:
		if d.Humans() == 0 {
			d.finish()
			return
		}
		d.settleBots()
		if d.roundDone() {
			d.nextRound()
		}
	}
}

// Standing is one seat's results so far.
type Standing struct {
	Seat   int `json:"seat"`
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

// Standings is every seat's win-loss record, by seat number. A bye
// counts for nothing.
func (d *Draft) Standings() []Standing {
	out := make([]Standing, len(d.seats))
	for i := range out {
		out[i].Seat = i
	}
	if d.play == nil {
		return out
	}
	for _, round := range d.play.rounds {
		for _, p := range round {
			if !p.Done || p.Bye() {
				continue
			}
			switch p.Winner {
			case NoSeat:
				out[p.A].Draws++
				out[p.B].Draws++
			case p.A:
				out[p.A].Wins++
				out[p.B].Losses++
			default:
				out[p.B].Wins++
				out[p.A].Losses++
			}
		}
	}
	return out
}

// SeatResult is one seat's match in one round, as a stored draft keeps
// it.
type SeatResult struct {
	Round int
	// Bye says the seat sat the round out; the opponent fields are
	// empty.
	Bye bool
	// OpponentBot and OpponentUser name the opponent as it was when the
	// round was paired.
	OpponentBot  bool
	OpponentUser uuid.UUID
	// Result is "win", "loss" or "draw", "" for a bye.
	Result              string
	GamesWon, GamesLost int
}

// Results is a seat's finished matches, in round order.
func (d *Draft) Results(seat int) []SeatResult {
	if d.play == nil {
		return nil
	}
	var out []SeatResult
	for r, round := range d.play.rounds {
		for _, p := range round {
			if !p.Done || !p.has(seat) {
				continue
			}
			res := SeatResult{Round: r + 1}
			switch {
			case p.Bye():
				res.Bye = true
			case p.A == seat:
				res.OpponentBot, res.OpponentUser = p.BotB, p.UserB
				res.GamesWon, res.GamesLost = p.WinsA, p.WinsB
			default:
				res.OpponentBot, res.OpponentUser = p.BotA, p.UserA
				res.GamesWon, res.GamesLost = p.WinsB, p.WinsA
			}
			if !p.Bye() {
				switch p.Winner {
				case NoSeat:
					res.Result = "draw"
				case seat:
					res.Result = "win"
				default:
					res.Result = "loss"
				}
			}
			out = append(out, res)
		}
	}
	return out
}

// WasPerson reports whether a person held the seat when the picks
// ended.
func (d *Draft) WasPerson(seat int) bool {
	return d.play != nil && seat >= 0 && seat < len(d.play.person) && d.play.person[seat]
}
