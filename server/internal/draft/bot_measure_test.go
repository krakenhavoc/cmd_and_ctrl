package draft

import (
	"flag"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// bot_measure_test.go is the draft bot's measure (ADR 0222 §7, #5439),
// run in CI on the fixture set: over MeasureDrafts seeded eight-bot
// drafts, every bot's deck is two colours (plus at most a splash) with a
// sane curve in at least MeasurePass of them, and a rated deck is worth
// more than a deck of random picks built from the same packs.
//
// The games-won half of the measure is in internal/botarena
// (draft_decks_test.go), which may hold the engine.

const (
	measureDrafts = 20
	measurePass   = 19
	measureSeed   = 5439
	// measureOutRate is the share of seats whose rated deck must be
	// worth more than the random-pick deck of the same seat.
	measureOutRate = 0.75
)

// -draft-wide runs the measure over this many drafts from other seeds,
// at eight, four and two seats, and logs the rates. It gates nothing: it
// is how the numbers in docs/bot.md were taken and how a change to a
// weight is checked against more than the twenty seeds CI holds.
var draftWide = flag.Int("draft-wide", 0, "also measure the draft bot over this many drafts of other seeds and log the rates")

// botDraft runs one whole draft with no person in it: every seat picks
// with `pickers[seat]`. It returns each seat's pool.
func botDraft(t testing.TB, fx *botFixture, mode Mode, seats int, seed uint64, picker Picker) [][]Card {
	t.Helper()
	d, err := New(Config{Name: "measure", Set: botFixtureSet, Mode: mode, Seats: seats, Seed: seed, Picker: picker}, t0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	host, err := d.AddSeat("host", uuid.Nil)
	if err != nil {
		t.Fatalf("AddSeat: %v", err)
	}
	if err := d.Start(host.ID, fx.src, t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// The host's seat is picked for by the same picker, as a seat whose
	// timer ran out is.
	for i := 0; d.State() == StatePicking; i++ {
		if i > 1000 {
			t.Fatal("the draft did not finish")
		}
		d.Expire(t0)
	}
	return d.Pools()
}

// deckReport is what the measure reads off a built deck.
type deckReport struct {
	Cards, Spells, Lands int
	Creatures, Removal   int
	// Colours is the colours of the deck's cards with how many cards
	// are of each, and Main the colours with more than SplashCards.
	Colours map[string]int
	Main    []string
	Splash  []string
	Curve   [7]int
	Average float64
	Rating  float64
	// LandColours is the colours the basic lands make.
	LandColours map[string]int
}

func reportDeck(r Rater, pool []Card, main []string, lands map[string]int) deckReport {
	rep := deckReport{Colours: map[string]int{}, LandColours: map[string]int{}}
	byID := map[string]Card{}
	for _, c := range pool {
		byID[c.ID] = c
	}
	total := 0
	for _, id := range main {
		rc := rateAll(r, []Card{byID[id]})[0]
		rep.Spells++
		rep.Rating += rc.r
		if rc.body() {
			rep.Creatures++
		}
		if rc.removal() {
			rep.Removal++
		}
		rep.Curve[rc.bucket()]++
		total += rc.f.ManaValue
		for _, col := range rc.c.Colors {
			rep.Colours[col]++
		}
	}
	for name, n := range lands {
		rep.Lands += n
		for letter, basic := range StandardBasics {
			if basic == name {
				rep.LandColours[letter] += n
			}
		}
	}
	rep.Cards = rep.Spells + rep.Lands
	if rep.Spells > 0 {
		rep.Average = float64(total) / float64(rep.Spells)
	}
	for i := range len(colourLetters) {
		col := colourLetters[i : i+1]
		switch n := rep.Colours[col]; {
		case n > SplashCards:
			rep.Main = append(rep.Main, col)
		case n > 0:
			rep.Splash = append(rep.Splash, col)
		}
	}
	return rep
}

// sane is the measure's reading of "two colours (plus at most a splash)
// with a sane curve".
func (rep deckReport) sane() (bool, string) {
	switch {
	case rep.Cards != DeckSize:
		return false, fmt.Sprintf("%d cards", rep.Cards)
	case rep.Lands < DeckLands || rep.Lands > DeckLands+1:
		return false, fmt.Sprintf("%d lands", rep.Lands)
	case len(rep.Main) > 2:
		return false, fmt.Sprintf("main colours %v", rep.Main)
	case len(rep.Main)+len(rep.Splash) > 3:
		return false, fmt.Sprintf("colours %v with splashes %v", rep.Main, rep.Splash)
	case rep.Creatures < 12:
		return false, fmt.Sprintf("%d creatures", rep.Creatures)
	case rep.Average < 2.0 || rep.Average > 3.8:
		return false, fmt.Sprintf("average mana value %.2f", rep.Average)
	case rep.Curve[1]+rep.Curve[2]+rep.Curve[3] < 9:
		return false, fmt.Sprintf("only %d cards of three mana or less", rep.Curve[1]+rep.Curve[2]+rep.Curve[3])
	case rep.Curve[5]+rep.Curve[6] > 8:
		return false, fmt.Sprintf("%d cards of five mana or more", rep.Curve[5]+rep.Curve[6])
	}
	return true, ""
}

// measured is one run of the measure.
type measured struct {
	drafts, saneDrafts int
	decks, saneDecks   int
	// outRated is the seats whose bot deck is worth more, by total
	// rating, than the random-pick deck of the same seat and seed.
	outRated            int
	botRating, rndTotal float64
	why                 []string
}

// measure drafts `drafts` times with eight bots and, from the same
// seeds and so the same packs, eight random pickers, and builds every
// pool with BuildDeck.
func measure(t testing.TB, fx *botFixture, mode Mode, seats, drafts int, seed0, step uint64) measured {
	t.Helper()
	var m measured
	for i := range drafts {
		seed := seed0 + uint64(i)*step
		pools := botDraft(t, fx, mode, seats, seed, BotPicker{Seed: seed, Rater: &fx.rater})
		random := botDraft(t, fx, mode, seats, seed, RandomPicker{Seed: seed})
		sane := true
		for seat, pool := range pools {
			main, lands := fx.rater.BuildDeck(pool, nil, seed)
			rep := reportDeck(fx.rater, pool, main, lands)
			rmain, rlands := fx.rater.BuildDeck(random[seat], nil, seed)
			rnd := reportDeck(fx.rater, random[seat], rmain, rlands)
			m.decks++
			if ok, why := rep.sane(); ok {
				m.saneDecks++
			} else {
				sane = false
				m.why = append(m.why, fmt.Sprintf("seed %d seat %d: %s (colours %v, splash %v, curve %v)", seed, seat, why, rep.Main, rep.Splash, rep.Curve[1:]))
			}
			if rep.Rating > rnd.Rating {
				m.outRated++
			}
			m.botRating += rep.Rating
			m.rndTotal += rnd.Rating
		}
		m.drafts++
		if sane {
			m.saneDrafts++
		}
	}
	return m
}

func (m measured) String() string {
	return fmt.Sprintf("every deck two colours with a sane curve in %d of %d drafts (%d of %d decks); "+
		"the rated deck out-rates the random-pick deck in %d of %d seats; mean deck rating %.1f against %.1f",
		m.saneDrafts, m.drafts, m.saneDecks, m.decks, m.outRated, m.decks,
		m.botRating/float64(m.decks), m.rndTotal/float64(m.decks))
}

// TestTheDraftBotMeasure is ADR 0222 §7's measure, the half that needs
// no engine: over twenty seeded eight-bot drafts of the fixture set,
// every bot's deck is two colours (plus at most a splash) with a sane
// curve in at least nineteen, and a rated deck is worth more than the
// random-pick deck built from the same packs. Both pick modes.
func TestTheDraftBotMeasure(t *testing.T) {
	fx := loadBotFixture(t)
	for _, mode := range []Mode{PickOne, PickTwo} {
		m := measure(t, fx, mode, 8, measureDrafts, measureSeed, 1)
		t.Logf("%s, 8 bots, seeds %d to %d: %s", mode, measureSeed, measureSeed+measureDrafts-1, m)
		for _, why := range m.why {
			t.Logf("%s: %s", mode, why)
		}
		if m.saneDrafts < measurePass {
			t.Errorf("%s: every deck is sane in %d of %d drafts, want at least %d", mode, m.saneDrafts, m.drafts, measurePass)
		}
		if float64(m.outRated) < measureOutRate*float64(m.decks) {
			t.Errorf("%s: the rated deck out-rates the random-pick deck in %d of %d seats, want at least %.0f%%",
				mode, m.outRated, m.decks, 100*measureOutRate)
		}
		if m.botRating <= m.rndTotal {
			t.Errorf("%s: the rated decks are worth %.1f in all and the random-pick decks %.1f", mode, m.botRating, m.rndTotal)
		}
	}
	if *draftWide > 0 {
		for _, mode := range []Mode{PickOne, PickTwo} {
			for _, seats := range []int{8, 4, 2} {
				t.Logf("wide: %s, %d bots, %d drafts: %s", mode, seats, *draftWide, measure(t, fx, mode, seats, *draftWide, 90000, 7))
			}
		}
	}
}

// What "sane" refuses, so the measure is not passed by a reading that
// accepts anything.
func TestTheMeasureRefusesABadDeck(t *testing.T) {
	good := deckReport{Cards: 40, Spells: 23, Lands: 17, Creatures: 15, Main: []string{"W", "R"},
		Curve: [7]int{0, 2, 6, 5, 4, 4, 2}, Average: 3.1}
	if ok, why := good.sane(); !ok {
		t.Fatalf("a good deck is refused: %s", why)
	}
	for name, change := range map[string]func(*deckReport){
		"39 cards":          func(r *deckReport) { r.Cards = 39 },
		"three colours":     func(r *deckReport) { r.Main = []string{"W", "U", "R"} },
		"two splashes":      func(r *deckReport) { r.Splash = []string{"U", "G"} },
		"16 lands":          func(r *deckReport) { r.Lands = 16 },
		"19 lands":          func(r *deckReport) { r.Lands = 19 },
		"nine creatures":    func(r *deckReport) { r.Creatures = 9 },
		"a heavy deck":      func(r *deckReport) { r.Average = 4.1 },
		"no early plays":    func(r *deckReport) { r.Curve = [7]int{0, 1, 3, 4, 6, 5, 4} },
		"a top-heavy curve": func(r *deckReport) { r.Curve = [7]int{0, 2, 5, 4, 3, 5, 4} },
	} {
		bad := good
		change(&bad)
		if ok, _ := bad.sane(); ok {
			t.Errorf("%s passes as sane", name)
		}
	}
}
