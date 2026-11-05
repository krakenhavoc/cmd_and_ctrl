package gamestats

// sampler.go is ADR 0217 §1's turn sampler: it reads a replay's frames
// once, oldest first, and keeps for each seat's own turns the facts §4
// needs, without holding a frame.
//
// For each turn it keeps a sample at the start of the active seat's
// precombat main phase (the first frame of the turn at or after it, so
// a commit that skipped through main still gives one), the turn's last
// frame, a sample at each declare_attackers step, and whether a land was
// in the seat's hand on any frame. It also keeps every card id that was
// ever in each seat's hand, from the kept hand on.
//
// An undo rewinds the game, and with it the log's Seq numbers: a frame
// whose newest log Seq is below the one before, or whose turn went
// backwards, is a rewind. Every sample is stamped with the newest Seq of
// the frame it came from, and a rewind to Seq S drops every sample
// stamped above S, so an undone draw or land play counts as not having
// happened. The frame itself is then sampled as any other.

import (
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Frames yields a game's omniscient views (a replay's frames), oldest
// first, until yield returns false, and reports whether the replay was
// read whole: every line decoded and the last frame is the ended game.
type Frames func(yield func(*protocol.GameView) bool) (complete bool)

// The game thresholds of ADR 0217 §4. Version-2 constants.
const (
	// ScrewGameTurns screw turns among the first ScrewGameWindow own
	// turns make a screwed game.
	ScrewGameTurns  = 2
	ScrewGameWindow = 6
	// FloodGameTurns flood turns make a flooded game.
	FloodGameTurns = 3
	// ColourScrewGameTurns colour-screw turns make a colour-screwed game.
	ColourScrewGameTurns = 2
)

// mainSample is a seat's state at the start of its precombat main.
type mainSample struct {
	at                        uint64
	handSize                  int
	landsInHand, spellsInHand int
	landsControlled           int
	manaSources               int
	cantPlayLands             bool
	colourScrew               bool
	// held is each card name in hand, and whether one copy of it was
	// castable by amount and colour.
	held map[string]bool
}

// lastSample is the turn's last frame, so far.
type lastSample struct {
	at                        uint64
	landsPlayed               int
	landsInHand, spellsInHand int
}

type attackSample struct {
	at                  uint64
	attackers, eligible int
}

// turnSample is one turn, of its active seat.
type turnSample struct {
	turn, seat int
	at         uint64 // the turn's first frame
	main       *mainSample
	last       *lastSample
	// landSeenAt is the stamp of the first frame with a land in the
	// seat's hand this turn; landSeen is false when there was none.
	landSeen   bool
	landSeenAt uint64
	// attacks is one sample per combat phase, keyed by its phase id.
	attacks map[int]attackSample
}

// seenCard is one card instance that was in a seat's hand.
type seenCard struct {
	at        uint64
	name      string
	firstTurn int
	opening   bool
}

type sampler struct {
	idx    CardIndex
	turns  map[int]*turnSample
	seen   map[int]map[string]*seenCard // seat → instance id → card
	frames int
	// prevSeq and prevTurn are the previous frame's newest log Seq and
	// turn, for spotting a rewind.
	prevSeq  uint64
	prevTurn int
	// opening is the seats' hands at the last frame before turn 1, and
	// openingDone is set once the kept hands are recorded.
	opening     map[int][]protocol.CardView
	openingDone bool
}

func newSampler(idx CardIndex) *sampler {
	return &sampler{idx: idx, turns: map[int]*turnSample{}, seen: map[int]map[string]*seenCard{}}
}

// newestSeq is the newest log Seq a frame carries.
func newestSeq(v *protocol.GameView) uint64 {
	var hi uint64
	for _, e := range v.Log {
		hi = max(hi, e.Seq)
	}
	return hi
}

// rewind drops every sample stamped above seq.
func (s *sampler) rewind(seq uint64) {
	for t, ts := range s.turns {
		if ts.at > seq {
			delete(s.turns, t)
			continue
		}
		if ts.main != nil && ts.main.at > seq {
			ts.main = nil
		}
		if ts.last != nil && ts.last.at > seq {
			ts.last = nil
		}
		if ts.landSeen && ts.landSeenAt > seq {
			ts.landSeen = false
		}
		for k, a := range ts.attacks {
			if a.at > seq {
				delete(ts.attacks, k)
			}
		}
	}
	for _, cards := range s.seen {
		for id, c := range cards {
			if c.at > seq {
				delete(cards, id)
			}
		}
	}
}

// stepAtOrAfterMain reports a step of the turn at or after the start of
// its precombat main phase.
func stepAtOrAfterMain(step string) bool {
	switch step {
	case "", "untap", "upkeep", "draw":
		return false
	}
	return true
}

// observe reads one frame.
func (s *sampler) observe(v *protocol.GameView) {
	s.frames++
	seq, turn := newestSeq(v), v.Turn.Seq
	if s.frames > 1 && (seq < s.prevSeq || turn < s.prevTurn) {
		s.rewind(seq)
	}
	s.prevSeq, s.prevTurn = seq, turn

	if turn < 1 {
		// Before the first turn: the hands being kept.
		s.opening = map[int][]protocol.CardView{}
		for i := range v.Seats {
			p := &v.Seats[i]
			s.opening[p.Seat] = append([]protocol.CardView(nil), p.Hand.Cards...)
		}
		return
	}
	if !s.openingDone {
		s.openingDone = true
		if s.opening == nil {
			// No frame before turn 1: the first frame's hands are the
			// kept ones.
			s.opening = map[int][]protocol.CardView{}
			for i := range v.Seats {
				p := &v.Seats[i]
				s.opening[p.Seat] = p.Hand.Cards
			}
		}
		for seat, hand := range s.opening {
			for _, c := range hand {
				s.see(seat, &c, 0, true, seq)
			}
		}
		s.opening = nil
	}
	for i := range v.Seats {
		p := &v.Seats[i]
		if p.Eliminated {
			continue
		}
		for j := range p.Hand.Cards {
			s.see(p.Seat, &p.Hand.Cards[j], turn, false, seq)
		}
	}

	if v.Turn.BetweenTurns {
		return
	}
	var active *protocol.PlayerView
	for i := range v.Seats {
		if v.Seats[i].Seat == v.Turn.ActiveSeat {
			active = &v.Seats[i]
		}
	}
	if active == nil || active.Eliminated {
		return
	}
	ts := s.turns[turn]
	if ts == nil {
		ts = &turnSample{turn: turn, seat: active.Seat, at: seq, attacks: map[int]attackSample{}}
		s.turns[turn] = ts
	}
	lands, spells, anyLand := handCounts(active.Hand.Cards)
	if anyLand && !ts.landSeen {
		ts.landSeen, ts.landSeenAt = true, seq
	}
	ts.last = &lastSample{at: seq, landsPlayed: active.LandsPlayedThisTurn, landsInHand: lands, spellsInHand: spells}
	if ts.main == nil && stepAtOrAfterMain(v.Turn.Step) {
		ts.main = s.mainOf(v, active, seq)
	}
	if v.Turn.Step == "declare_attackers" {
		key := v.Turn.PhaseID
		if key == 0 {
			key = v.Turn.PhaseOrdinal
		}
		a := attackSample{at: seq}
		for j := range v.Battlefield.Cards {
			c := &v.Battlefield.Cards[j]
			if c.Controller != active.ID || c.PhasedOut || !isCreature(c) {
				continue
			}
			attacking := c.AttackingTarget != ""
			if attacking {
				a.attackers++
			}
			if !c.SummoningSick && (attacking || !c.Tapped) {
				a.eligible++
			}
		}
		ts.attacks[key] = a
	}
}

func isCreature(c *protocol.CardView) bool {
	front, _, _ := strings.Cut(c.TypeLine, " // ")
	return strings.Contains(front, "Creature")
}

// handCounts: lands and spells for flood (a land-backed modal card is
// neither), and whether any card could be played as a land, for screw.
func handCounts(hand []protocol.CardView) (lands, spells int, anyLand bool) {
	for i := range hand {
		switch landKindOf(&hand[i]) {
		case landCard:
			lands++
			anyLand = true
		case landBackCard:
			anyLand = true
		default:
			spells++
		}
	}
	return lands, spells, anyLand
}

func (s *sampler) see(seat int, c *protocol.CardView, turn int, opening bool, seq uint64) {
	if c.InstanceID == "" || c.Name == "" || c.IsToken {
		return
	}
	m := s.seen[seat]
	if m == nil {
		m = map[string]*seenCard{}
		s.seen[seat] = m
	}
	if _, ok := m[c.InstanceID]; ok {
		return
	}
	m[c.InstanceID] = &seenCard{at: seq, name: c.Name, firstTurn: turn, opening: opening}
}

// mainOf samples the active seat at the start of its main phase.
func (s *sampler) mainOf(v *protocol.GameView, p *protocol.PlayerView, seq uint64) *mainSample {
	m := &mainSample{at: seq, handSize: len(p.Hand.Cards), cantPlayLands: p.CantPlayLands != "", held: map[string]bool{}}
	m.landsInHand, m.spellsInHand, _ = handCounts(p.Hand.Cards)
	var sources []colours
	for j := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[j]
		if c.Controller != p.ID || c.PhasedOut {
			continue
		}
		front, _, _ := strings.Cut(c.TypeLine, " // ")
		if isLandFace(front) {
			m.landsControlled++
		}
		if c.Tapped || (isCreature(c) && c.SummoningSick) {
			continue
		}
		if made := producedBy(c, s.idx); made != 0 {
			sources = append(sources, made)
		}
	}
	m.manaSources = len(sources)
	for j := range p.Hand.Cards {
		c := &p.Hand.Cards[j]
		if c.Name == "" {
			continue
		}
		if landKindOf(c) == landCard {
			continue
		}
		byAmount, byColour := castable(c.ManaCost, sources)
		if byAmount && !byColour {
			m.colourScrew = true
		}
		m.held[c.Name] = m.held[c.Name] || byColour
	}
	return m
}

// TurnStat is one row of seat_turn_stats: one own turn of a seat.
type TurnStat struct {
	OwnTurn int
	Turn    int
	// LandPlayed is the land plays this turn (at its last frame), and
	// LandDrops the plays so far this game, this turn included.
	LandPlayed int
	LandDrops  int
	// At the start of main.
	LandsControlled int
	ManaSources     int
	HandSize        int
	LandsInHand     int
	SpellsInHand    int

	Missed      bool
	Screw       bool
	Flood       bool
	ColourScrew bool

	Attackers int
	Eligible  int
	// From the log.
	DamageToPlayers int
	SpellsCast      int
}

// SeenCard is one row of seat_card_seen: a card name that was in a
// seat's hand.
type SeenCard struct {
	Name          string
	Copies        int
	InOpeningHand bool
	// FirstSeenTurn is the game turn it was first in hand; 0 for the
	// opening hand.
	FirstSeenTurn   int
	HeldTurns       int
	AffordableTurns int
	Cast            bool
}

// Play is a seat's frame-derived record (ADR 0217 §4). It is nil when
// the game's replay was not read.
type Play struct {
	OwnTurns          int
	LandDropsHit      int
	LandDropsMissed   int
	ScrewTurns        int
	FloodTurns        int
	ColourScrewTurns  int
	ScrewGame         bool
	FloodGame         bool
	ColourScrewGame   bool
	CreaturesEligible int
	CreaturesAttacked int
	Turns             []TurnStat
	Seen              []SeenCard
}

// play folds the samples into each seat's Play. perTurn gives the log's
// damage to players and spells cast for a seat and turn; cast is the
// card names each seat cast.
func (s *sampler) play(seats []int, perTurn func(seat, turn int) (damage, spells int), cast map[int]map[string]*Card) map[int]*Play {
	out := make(map[int]*Play, len(seats))
	for _, n := range seats {
		out[n] = &Play{Turns: []TurnStat{}, Seen: []SeenCard{}}
	}
	order := make([]int, 0, len(s.turns))
	for t := range s.turns {
		order = append(order, t)
	}
	sort.Ints(order)
	held := map[int]map[string]*SeenCard{}
	for _, t := range order {
		ts := s.turns[t]
		p := out[ts.seat]
		if p == nil || ts.main == nil {
			continue // a turn that never reached its main phase
		}
		p.OwnTurns++
		row := TurnStat{
			OwnTurn: p.OwnTurns, Turn: t,
			LandsControlled: ts.main.landsControlled, ManaSources: ts.main.manaSources,
			HandSize: ts.main.handSize, LandsInHand: ts.main.landsInHand, SpellsInHand: ts.main.spellsInHand,
			ColourScrew: ts.main.colourScrew,
		}
		if ts.last != nil {
			row.LandPlayed = ts.last.landsPlayed
		}
		prev := 0
		if k := len(p.Turns); k > 0 {
			prev = p.Turns[k-1].LandDrops
		}
		row.LandDrops = prev + row.LandPlayed
		if row.LandPlayed == 0 && !ts.main.cantPlayLands {
			row.Missed = true
			row.Screw = !ts.landSeen
		}
		if row.LandPlayed >= 1 && ts.last != nil && ts.last.landsInHand > ts.last.spellsInHand {
			row.Flood = true
		}
		for _, a := range ts.attacks {
			row.Attackers += a.attackers
			row.Eligible += a.eligible
		}
		row.DamageToPlayers, row.SpellsCast = perTurn(ts.seat, t)

		if row.Missed {
			p.LandDropsMissed++
		} else if row.LandPlayed >= 1 {
			p.LandDropsHit++
		}
		if row.Screw {
			p.ScrewTurns++
		}
		if row.Flood {
			p.FloodTurns++
		}
		if row.ColourScrew {
			p.ColourScrewTurns++
		}
		p.CreaturesAttacked += row.Attackers
		p.CreaturesEligible += row.Eligible
		p.Turns = append(p.Turns, row)

		if held[ts.seat] == nil {
			held[ts.seat] = map[string]*SeenCard{}
		}
		for name, affordable := range ts.main.held {
			c := held[ts.seat][name]
			if c == nil {
				c = &SeenCard{Name: name}
				held[ts.seat][name] = c
			}
			c.HeldTurns++
			if affordable {
				c.AffordableTurns++
			}
		}
	}
	for _, n := range seats {
		p := out[n]
		early := 0
		for _, r := range p.Turns {
			if r.OwnTurn <= ScrewGameWindow && r.Screw {
				early++
			}
		}
		p.ScrewGame = early >= ScrewGameTurns
		p.FloodGame = p.FloodTurns >= FloodGameTurns
		p.ColourScrewGame = p.ColourScrewTurns >= ColourScrewGameTurns

		byName := map[string]*SeenCard{}
		for _, c := range s.seen[n] {
			sc := byName[c.name]
			if sc == nil {
				sc = &SeenCard{Name: c.name, FirstSeenTurn: c.firstTurn}
				if h := held[n][c.name]; h != nil {
					sc.HeldTurns, sc.AffordableTurns = h.HeldTurns, h.AffordableTurns
				}
				sc.Cast = cast[n][c.name] != nil
				byName[c.name] = sc
			}
			sc.Copies++
			sc.InOpeningHand = sc.InOpeningHand || c.opening
			sc.FirstSeenTurn = min(sc.FirstSeenTurn, c.firstTurn)
		}
		for _, sc := range byName {
			p.Seen = append(p.Seen, *sc)
		}
		sort.Slice(p.Seen, func(a, b int) bool { return p.Seen[a].Name < p.Seen[b].Name })
	}
	return out
}
