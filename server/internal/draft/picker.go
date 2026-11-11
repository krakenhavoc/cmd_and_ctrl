package draft

import "math/rand/v2"

// PickRequest is what a Picker is asked: the pack the seat holds, what
// it has drafted so far and the pick mode.
type PickRequest struct {
	// Pack is the pack the seat holds. Never empty.
	Pack []Card
	// Picked is every card the seat has drafted, in the order taken.
	Picked []Card
	Mode   Mode
	// Take is how many cards this pick takes: Mode.PerPick(), or fewer
	// when the pack is nearly empty. Always at least 1 and at most
	// len(Pack).
	Take int

	// Where the pick falls. A picker may ignore all of it.
	//
	// Seat is the seat's number from 0 and Seats the size of the draft.
	// Round and Pick count from 1.
	Seat, Seats int
	Round, Pick int
}

// Picker chooses a pick for a seat no person is choosing for: a bot
// seat, a seat whose pick timer ran out, and a seat that is away.
//
// It returns the IDs of exactly Take different cards of Pack. Anything
// else is overruled with the first cards of the pack.
//
// This interface is the whole seam for a draft bot. BotPicker (#5439)
// is the bot that rates cards and drafts toward a deck; RandomPicker is
// the placeholder it replaced, kept as the baseline the bot is measured
// against. A Picker is called with the draft's lock held, so it must
// not block and must not call back into the draft.
type Picker interface {
	Pick(PickRequest) []string
}

// RandomPicker takes cards of the pack at random. It was the placeholder
// bot of slice 4 and is now the baseline of the draft bot's measure
// (bot_measure_test.go): a deck of random picks from the same packs.
//
// Its choice depends only on the seed and on where the pick falls (the
// seat, the round and the pick number), never on the order picks were
// asked for. So a seeded draft replays the same way whatever the timing
// of its seats.
type RandomPicker struct {
	Seed uint64
}

// Pick takes Take cards of the pack at random.
func (p RandomPicker) Pick(req PickRequest) []string {
	// One PCG stream per pick. The high bit keeps every one clear of
	// packStream, so the packs and the picks never share a sequence.
	stream := 1<<63 | uint64(req.Seat)<<32 | uint64(req.Round)<<16 | uint64(req.Pick)
	rng := rand.New(rand.NewPCG(p.Seed, stream))
	order := rng.Perm(len(req.Pack))
	n := min(max(req.Take, 0), len(order))
	out := make([]string, 0, n)
	for _, i := range order[:n] {
		out = append(out, req.Pack[i].ID)
	}
	return out
}

// BotPicker is the draft bot (ADR 0222 §7, #5439). It takes the best
// card early, leans toward the colours it has taken with a commitment
// that grows with each pick and is firm after about the first pack,
// weighs what its deck still needs (curve, creatures, removal), and in
// pick two chooses the best pair, not the best card twice.
//
// Like RandomPicker it keeps no state: a pick is a function of the
// seed, where the pick falls and the cards in the request. So a seeded
// draft replays whatever the timing of its seats, and the bot that
// picks for a person whose timer ran out picks from that person's own
// picks so far. It never blocks: it reads the request and the rater's
// in-memory lookups.
type BotPicker struct {
	Seed uint64
	// Rater rates the cards. Nil is the package's rater over the
	// server's card index (UseCards).
	Rater *Rater
}

// The weights of a pick, all of them. They are in the rating's points
// (rating.go).
const (
	// CommitPacks is how many packs' worth of picks it takes for the
	// colour commitment to become firm: one.
	CommitPacks = 1.0
	// OffColourPenalty is what a fully committed bot takes from a card
	// outside its two colours. It is more than any rating, so a firm bot
	// takes an off-colour card only when the pack has nothing else.
	OffColourPenalty = 6.0
	// ColourCountsAbove is the rating above which a picked card pulls
	// the bot toward its colour, by the amount it is above. A filler
	// card says little about which colours the deck is.
	ColourCountsAbove = 1.5
	// GoldBeforeCommit is taken from a card of two or more colours while
	// the bot is uncommitted, scaled by how uncommitted it is: such a
	// card fits one deck in ten.
	GoldBeforeCommit = 0.6

	// NeedCreature is what a creature gains when the deck is short of
	// them, and DeckCreatures the count a finished deck wants.
	NeedCreature  = 1.0
	DeckCreatures = 15
	// NeedCreatureFull is how many creatures behind its pace a deck is
	// when a creature gains the whole of NeedCreature, and
	// MaxNeedCreature how many times that a deck far behind can add.
	NeedCreatureFull = 1.5
	MaxNeedCreature  = 4.0
	// NeedAhead is how far ahead of the draft a deck wants its creatures
	// and removal: the last picks of a pack are what seven others left,
	// so a deck that is on pace only at the end is short.
	NeedAhead = 1.3
	// NeedRemoval is what removal gains while the deck is short of it,
	// and DeckRemoval the count it wants.
	NeedRemoval = 0.4
	DeckRemoval = 4
	// CurveNeed is what a cheap card gains while its slot of the curve
	// is behind, and CurveGlut what each card over a full slot costs,
	// up to CurveGlutMax.
	CurveNeed    = 0.3
	CurveGlut    = 0.5
	CurveGlutMax = 1.5

	// PickNoise is the half-width of the seeded noise added to every
	// card's score, so eight bots do not draft identically.
	PickNoise = 0.2
	// PairSameColour is what the second card of a pick-two pair gains
	// for sharing a colour with the first while the bot is uncommitted.
	PairSameColour = 0.4
)

// curveIdeal is how many nonland cards of each mana value a deck of
// DeckSpells wants: index 1 is one mana or less, 6 is six or more.
var curveIdeal = [7]int{0, 2, 6, 5, 4, 3, 2}

// Pick makes the bot's pick.
func (p BotPicker) Pick(req PickRequest) []string {
	take := min(max(req.Take, 0), len(req.Pack))
	if take == 0 {
		return nil
	}
	rater := defaultRater()
	if p.Rater != nil {
		rater = *p.Rater
	}
	per := max(req.Mode.PerPick(), 1)
	// What the pack held when it was opened: the picks before this one
	// each took `per` cards from it.
	packSize := len(req.Pack) + max(req.Pick-1, 0)*per
	have := rateAll(rater, req.Picked)
	pack := rateAll(rater, req.Pack)
	noise := make([]float64, len(pack))
	for i, c := range pack {
		noise[i] = PickNoise * pickNoise(p.Seed, req, c.c.ID)
	}
	base := newPickState(have, packSize)

	if take == 2 && len(pack) >= 2 {
		return p.bestPair(base, have, pack, noise, packSize)
	}
	// One card, or the rare pick of three or more: the best card, then
	// the best given that one, and so on.
	chosen := make([]string, 0, take)
	used := make([]bool, len(pack))
	st := base
	for range take {
		best, bestScore := -1, 0.0
		for i, c := range pack {
			if used[i] {
				continue
			}
			if s := st.score(c) + noise[i]; best < 0 || s > bestScore {
				best, bestScore = i, s
			}
		}
		used[best] = true
		chosen = append(chosen, pack[best].c.ID)
		have = append(have, pack[best])
		st = newPickState(have, packSize)
	}
	return chosen
}

// bestPair is the pick-two choice: the two cards that are worth most
// together. The second card is scored as a pick made after the first,
// so two cards that pull the deck two ways are worth less than their
// ratings added, and two that share a colour are worth more.
func (p BotPicker) bestPair(base pickState, have, pack []rated, noise []float64, packSize int) []string {
	first := make([]float64, len(pack))
	after := make([]pickState, len(pack))
	for i, c := range pack {
		first[i] = base.score(c) + noise[i]
		after[i] = newPickState(append(have[:len(have):len(have)], c), packSize)
	}
	bi, bj, best := -1, -1, 0.0
	for i := range pack {
		for j := range pack {
			if i == j {
				continue
			}
			// i is taken first, j second.
			s := first[i] + after[i].score(pack[j]) + noise[j]
			if sharesColour(pack[i].c, pack[j].c) {
				s += PairSameColour * (1 - base.commit)
			}
			if bi < 0 || s > best {
				bi, bj, best = i, j, s
			}
		}
	}
	return []string{pack[bi].c.ID, pack[bj].c.ID}
}

func sharesColour(a, b Card) bool {
	for _, x := range a.Colors {
		for _, y := range b.Colors {
			if x == y {
				return true
			}
		}
	}
	return false
}

// rated is a card with what the bot knows of it.
type rated struct {
	c Card
	f Facts
	r float64
	// colours is the card's colours as indexes into colourLetters.
	colours []int
}

// removal reports whether the card answers a creature.
func (c rated) removal() bool {
	return c.f.Removal || c.f.Sweep || c.f.Damage >= 2
}

// body reports whether the card puts a creature on the battlefield: a
// creature, or a spell that makes creature tokens. It is what a deck's
// creature count counts.
func (c rated) body() bool { return c.f.Creature || c.f.Tokens > 0 }

// bucket is the card's slot of the curve: 1 for one mana or less, 6 for
// six or more.
func (c rated) bucket() int { return min(max(c.f.ManaValue, 1), 6) }

func rateAll(r Rater, cs []Card) []rated {
	out := make([]rated, 0, len(cs))
	for _, c := range cs {
		rc := rated{c: c, f: r.FactsOf(c), r: r.Rating(c)}
		for _, col := range c.Colors {
			if i := colourIndex(col); i >= 0 {
				rc.colours = append(rc.colours, i)
			}
		}
		out = append(out, rc)
	}
	return out
}

// colourLetters is the five colours in the order every tie is broken.
const colourLetters = "WUBRG"

func colourIndex(letter string) int {
	if len(letter) != 1 {
		return -1
	}
	for i := range len(colourLetters) {
		if colourLetters[i] == letter[0] {
			return i
		}
	}
	return -1
}

// pickState is what a bot has drafted, as the pick reads it.
type pickState struct {
	// commit is the colour commitment, 0 before the first pick and 1
	// once it is firm.
	commit float64
	// progress is how far through the draft the next pick is, 0 to 1.
	progress float64
	// weight is how much the picks pull toward each colour.
	weight [5]float64
	// first and second are the bot's two colours so far, -1 when it has
	// none.
	first, second int
	shape         deckShape
}

// deckShape counts what a deck in the making holds. creatures counts
// bodies (rated.body).
type deckShape struct {
	spells, creatures, removal int
	curve                      [7]int
}

// topEnd is the cards of TopEndManaValue or more.
func (s deckShape) topEnd() int { return s.curve[5] + s.curve[6] }

func (s *deckShape) add(c rated) {
	if c.f.Land {
		return
	}
	s.spells++
	if c.body() {
		s.creatures++
	}
	if c.removal() {
		s.removal++
	}
	s.curve[c.bucket()]++
}

// need is what a card gains or loses for what the deck still needs, at
// a point `progress` of the way to a finished deck.
func (s deckShape) need(c rated, progress float64) float64 {
	if c.f.Land {
		return 0
	}
	v := 0.0
	if c.body() {
		if short := float64(DeckCreatures)*min(progress*NeedAhead, 1) - float64(s.creatures); short > 0 {
			v += NeedCreature * min(short/NeedCreatureFull, MaxNeedCreature)
		}
	}
	if c.removal() && float64(s.removal) < float64(DeckRemoval)*min(progress*NeedAhead, 1) {
		v += NeedRemoval
	}
	b := c.bucket()
	ideal := curveIdeal[b]
	switch have := s.curve[b]; {
	case b >= TopEndManaValue && s.topEnd() >= MaxTopEnd:
		// The deck will not play another card this dear.
		v -= CurveGlutMax
	case have > ideal:
		v -= min(CurveGlut*float64(have-ideal), CurveGlutMax)
	case b <= 3 && float64(have) < float64(ideal)*progress:
		v += CurveNeed
	}
	return v
}

func newPickState(have []rated, packSize int) pickState {
	st := pickState{first: -1, second: -1}
	n := len(have)
	if packSize > 0 {
		st.commit = min(float64(n)/(CommitPacks*float64(packSize)), 1)
		st.progress = min(float64(n+1)/float64(Rounds*packSize), 1)
	}
	for _, c := range have {
		if c.r <= ColourCountsAbove || len(c.colours) == 0 {
			continue
		}
		pull := (c.r - ColourCountsAbove) / float64(len(c.colours))
		for _, i := range c.colours {
			st.weight[i] += pull
		}
	}
	for i, w := range st.weight {
		switch {
		case w <= 0:
		case st.first < 0 || w > st.weight[st.first]:
			st.first, st.second = i, st.first
		case st.second < 0 || w > st.weight[st.second]:
			st.second = i
		}
	}
	for _, c := range have {
		if st.affinity(c) >= 1 {
			st.shape.add(c)
		}
	}
	return st
}

// affinity is how well a card fits the bot's colours: 1 for a card it
// can play (colourless, or in its two colours), down to 0 for a colour
// it has taken nothing of. A bot with one colour so far is open to any
// second.
func (st pickState) affinity(c rated) float64 {
	if len(c.colours) == 0 || st.first < 0 {
		return 1
	}
	a := 1.0
	for _, i := range c.colours {
		a = min(a, st.colourAffinity(i))
	}
	return a
}

func (st pickState) colourAffinity(i int) float64 {
	if i == st.first || i == st.second || st.second < 0 {
		return 1
	}
	return min(st.weight[i]/st.weight[st.second], 1)
}

// score is what taking the card is worth now.
func (st pickState) score(c rated) float64 {
	a := st.affinity(c)
	s := c.r - st.commit*OffColourPenalty*(1-a)
	if len(c.colours) >= 2 {
		s -= GoldBeforeCommit * (1 - st.commit)
	}
	if a >= 1 {
		// Needs count for more as the colours settle: before that the
		// bot does not know which of its picks are its deck.
		s += st.commit * st.shape.need(c, st.progress)
	}
	return s
}

// pickNoise is a number in [-1, 1) that depends only on the seed, where
// the pick falls and the card, never on the order of the pack or on any
// other pick.
func pickNoise(seed uint64, req PickRequest, cardID string) float64 {
	h := seed ^ 0x5439_5439_5439_5439
	mix := func(v uint64) {
		h ^= v
		h *= 0x9E3779B97F4A7C15
		h ^= h >> 32
	}
	mix(uint64(req.Seat) + 1)
	mix(uint64(req.Round)<<16 | uint64(req.Pick))
	for i := range len(cardID) {
		mix(uint64(cardID[i]))
	}
	h ^= h >> 29
	h *= 0xBF58476D1CE4E5B9
	h ^= h >> 32
	return float64(h>>11)/float64(1<<52) - 1
}
