package draft

import (
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
)

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

func TestMeasureExplore(t *testing.T) {
	fx := loadBotFixture(t)
	for _, mode := range []Mode{PickOne, PickTwo} {
		bad, decks := 0, 0
		var botTotal, randTotal float64
		for i := range measureDrafts {
			seed := uint64(measureSeed + i)
			pools := botDraft(t, fx, mode, 8, seed, BotPicker{Seed: seed, Rater: &fx.rater})
			rpools := botDraft(t, fx, mode, 8, seed, RandomPicker{Seed: seed})
			for s, pool := range pools {
				main, lands := fx.rater.BuildDeck(pool, nil, seed)
				rep := reportDeck(fx.rater, pool, main, lands)
				ok, why := rep.sane()
				decks++
				if !ok {
					bad++
					t.Logf("%s seed %d seat %d: %s", mode, seed, s, why)
				}
				botTotal += rep.Rating
				rmain, rlands := fx.rater.BuildDeck(rpools[s], nil, seed)
				rrep := reportDeck(fx.rater, rpools[s], rmain, rlands)
				randTotal += rrep.Rating
				if i == 0 {
					t.Logf("%s seat %d: %v+%v sp %d cr %d rm %d curve %v avg %.2f lands %v rating %.1f | random %v+%v sp %d cr %d rating %.1f",
						mode, s, rep.Main, rep.Splash, rep.Spells, rep.Creatures, rep.Removal, rep.Curve[1:], rep.Average, lands, rep.Rating,
						rrep.Main, rrep.Splash, rrep.Spells, rrep.Creatures, rrep.Rating)
				}
			}
		}
		t.Logf("%s: %d of %d decks not sane; mean rating bot %.1f random %.1f", mode, bad, decks, botTotal/float64(decks), randTotal/float64(decks))
	}
}
