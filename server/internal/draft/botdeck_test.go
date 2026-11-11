package draft

import (
	"reflect"
	"strconv"
	"testing"
)

// checkLegal fails the test unless the deck meets the limited deck rule
// (CR 100.2b: at least 40 cards, of the player's product and basic
// lands; CR 905.1d: the pool plus any number of basic land cards): every
// main-deck card is a different card of the pool, the rest is basic
// lands the caller offered, and it is 40 cards in all.
func checkLegal(t testing.TB, pool []Card, basics BasicLands, main []string, lands map[string]int) {
	t.Helper()
	if len(basics) == 0 {
		basics = StandardBasics
	}
	inPool := map[string]Card{}
	for _, c := range pool {
		inPool[c.ID] = c
	}
	seen := map[string]bool{}
	for _, id := range main {
		c, ok := inPool[id]
		switch {
		case !ok:
			t.Errorf("main deck card %q is not in the pool", id)
		case seen[id]:
			t.Errorf("main deck card %q (%s) is played twice", id, c.Name)
		}
		seen[id] = true
	}
	offered := map[string]bool{}
	for _, name := range basics {
		offered[name] = true
	}
	total := len(main)
	for name, n := range lands {
		if !offered[name] {
			t.Errorf("the deck adds %d %q, which is not a basic land it was offered", n, name)
		}
		if n <= 0 {
			t.Errorf("the deck adds %d %q", n, name)
		}
		total += n
	}
	if total < 40 {
		t.Errorf("the deck is %d cards, below the limited minimum of 40 (CR 100.2b)", total)
	}
	if total != DeckSize {
		t.Errorf("the deck is %d cards, want exactly %d", total, DeckSize)
	}
	// In pool order.
	i := 0
	for _, c := range pool {
		if i < len(main) && main[i] == c.ID {
			i++
		}
	}
	if i != len(main) {
		t.Errorf("the main deck is not in pool order")
	}
}

// Every deck BuildDeck makes of a drafted pool is a legal limited deck,
// for bots and for pools nobody chose well.
func TestBuildDeckIsALegalLimitedDeck(t *testing.T) {
	fx := loadBotFixture(t)
	for _, mode := range []Mode{PickOne, PickTwo} {
		for seed := uint64(1); seed <= 6; seed++ {
			pools := botDraft(t, fx, mode, 8, seed, BotPicker{Seed: seed, Rater: &fx.rater})
			pools = append(pools, botDraft(t, fx, mode, 3, seed, RandomPicker{Seed: seed})...)
			for _, pool := range pools {
				main, lands := fx.rater.BuildDeck(pool, nil, seed)
				checkLegal(t, pool, nil, main, lands)
				for _, id := range main {
					for _, c := range pool {
						if c.ID == id && (c.Basic || fx.rater.FactsOf(c).Land) {
							t.Errorf("%s is in the main deck: the deck's lands are basics", c.Name)
						}
					}
				}
			}
		}
	}
}

func repeatCards(fx *botFixture, t testing.TB, prefix string, n int, name string) []Card {
	out := make([]Card, 0, n)
	for i := range n {
		out = append(out, fx.card(t, name, prefix+strconv.Itoa(i)))
	}
	return out
}

// The shape of the deck: two colours, 23 nonland cards and 17 lands (22
// and 18 when it is heavy), and the lands split by the coloured mana of
// its cards.
func TestBuildDeckColoursCountAndLandSplit(t *testing.T) {
	fx := loadBotFixture(t)

	// A white deck with a little red: fifteen cards of {1}{W}-like
	// costs, eight of red, and a pile of green and blue it must leave.
	var pool []Card
	pool = append(pool, repeatCards(fx, t, "w", 8, "C2 Fixture White Cub")...)    // {1}{W}
	pool = append(pool, repeatCards(fx, t, "a", 4, "Pacifism")...)                // {1}{W}
	pool = append(pool, repeatCards(fx, t, "s", 3, "C2 Fixture White Sentry")...) // {2}{W}
	pool = append(pool, repeatCards(fx, t, "r", 5, "C2 Fixture Red Scout")...)    // {2}{R}
	pool = append(pool, repeatCards(fx, t, "l", 3, "Lightning Strike")...)        // {1}{R}
	pool = append(pool, repeatCards(fx, t, "g", 6, "Giant Growth")...)
	pool = append(pool, repeatCards(fx, t, "u", 6, "Cancel")...)
	pool = append(pool, repeatCards(fx, t, "p", 3, "Plains")...)
	main, lands := fx.rater.BuildDeck(pool, nil, 1)
	checkLegal(t, pool, nil, main, lands)
	if len(main) != DeckSpells {
		t.Errorf("%d nonland cards, want %d", len(main), DeckSpells)
	}
	rep := reportDeck(fx.rater, pool, main, lands)
	if !reflect.DeepEqual(rep.Main, []string{"W", "R"}) || len(rep.Splash) != 0 {
		t.Errorf("the deck's colours are %v with splash %v, want white and red", rep.Main, rep.Splash)
	}
	// Fifteen white symbols and eight red: 17 lands split 11 and 6.
	if !reflect.DeepEqual(lands, map[string]int{"Plains": 11, "Mountain": 6}) {
		t.Errorf("the lands are %v, want 11 Plains and 6 Mountains", lands)
	}

	// A heavy deck plays 22 and 18.
	var heavy []Card
	heavy = append(heavy, repeatCards(fx, t, "a", 6, "C2 Fixture Green Giant")...)
	heavy = append(heavy, repeatCards(fx, t, "b", 6, "C2 Fixture Green Guard")...)
	heavy = append(heavy, repeatCards(fx, t, "c", 6, "C2 Fixture Green Brute")...)
	heavy = append(heavy, repeatCards(fx, t, "d", 6, "C2 Fixture Black Brute")...)
	heavy = append(heavy, repeatCards(fx, t, "e", 4, "C2 Fixture Black Scout")...)
	main, lands = fx.rater.BuildDeck(heavy, nil, 1)
	checkLegal(t, heavy, nil, main, lands)
	n := 0
	for _, k := range lands {
		n += k
	}
	if len(main) != DeckSpells-1 || n != DeckLands+1 {
		t.Errorf("a heavy deck plays %d spells and %d lands, want %d and %d", len(main), n, DeckSpells-1, DeckLands+1)
	}

	// One colour: every land is that colour's basic.
	mono := repeatCards(fx, t, "m", 30, "C2 Fixture Blue Cub")
	main, lands = fx.rater.BuildDeck(mono, nil, 1)
	checkLegal(t, mono, nil, main, lands)
	if !reflect.DeepEqual(lands, map[string]int{"Island": 17}) {
		t.Errorf("a mono-blue deck's lands are %v", lands)
	}

	// The basics are the ones the caller offers, by the names it gives.
	snow := BasicLands{"W": "Snow-Covered Plains", "R": "Snow-Covered Mountain"}
	main, lands = fx.rater.BuildDeck(pool, snow, 1)
	checkLegal(t, pool, snow, main, lands)
	if !reflect.DeepEqual(lands, map[string]int{"Snow-Covered Plains": 11, "Snow-Covered Mountain": 6}) {
		t.Errorf("with snow basics the lands are %v", lands)
	}
}

// The curve check: the deck takes its cheap cards before a seventh card
// of five mana or more, though each of those rates higher, and keeps its
// creature count.
func TestBuildDeckChecksTheCurve(t *testing.T) {
	fx := loadBotFixture(t)
	var pool []Card
	pool = append(pool, repeatCards(fx, t, "g", 12, "C2 Fixture Green Giant")...) // six mana, rated above the Cub
	pool = append(pool, repeatCards(fx, t, "c", 10, "C2 Fixture Green Cub")...)
	pool = append(pool, repeatCards(fx, t, "s", 8, "C2 Fixture Green Scout")...)
	pool = append(pool, repeatCards(fx, t, "b", 10, "Giant Growth")...)
	if fx.rating(t, "C2 Fixture Green Giant") <= fx.rating(t, "C2 Fixture Green Cub") {
		t.Fatal("the test needs the Giant to rate above the Cub")
	}
	main, lands := fx.rater.BuildDeck(pool, nil, 1)
	checkLegal(t, pool, nil, main, lands)
	rep := reportDeck(fx.rater, pool, main, lands)
	if top := rep.Curve[5] + rep.Curve[6]; top != MaxTopEnd {
		t.Errorf("%d cards of five mana or more, want the cap of %d", top, MaxTopEnd)
	}
	if rep.Creatures < MinDeckCreatures {
		t.Errorf("%d creatures, want at least %d", rep.Creatures, MinDeckCreatures)
	}

	// Removal-heavy: the last slots are kept for creatures.
	pool = nil
	pool = append(pool, repeatCards(fx, t, "m", 20, "Murder")...)
	pool = append(pool, repeatCards(fx, t, "c", 14, "C2 Fixture Black Cub")...)
	main, lands = fx.rater.BuildDeck(pool, nil, 1)
	checkLegal(t, pool, nil, main, lands)
	if rep := reportDeck(fx.rater, pool, main, lands); rep.Creatures != MinDeckCreatures {
		t.Errorf("%d creatures beside twenty Murders, want %d", rep.Creatures, MinDeckCreatures)
	}
}

// A third colour is a light splash, and only when the pool forces it.
func TestBuildDeckSplashesOnlyWhenForced(t *testing.T) {
	fx := loadBotFixture(t)
	var pool []Card
	pool = append(pool, repeatCards(fx, t, "w", 10, "C2 Fixture White Cub")...)
	pool = append(pool, repeatCards(fx, t, "u", 8, "C2 Fixture Blue Scout")...)
	pool = append(pool, repeatCards(fx, t, "r", 5, "Lightning Strike")...) // one red symbol each
	pool = append(pool, repeatCards(fx, t, "f", 2, "Fire Elemental")...)   // two: too heavy to splash
	main, lands := fx.rater.BuildDeck(pool, nil, 1)
	checkLegal(t, pool, nil, main, lands)
	rep := reportDeck(fx.rater, pool, main, lands)
	if !reflect.DeepEqual(rep.Main, []string{"W", "U"}) || !reflect.DeepEqual(rep.Splash, []string{"R"}) {
		t.Fatalf("colours %v, splash %v: want white and blue, splashing red", rep.Main, rep.Splash)
	}
	if rep.Colours["R"] != SplashCards {
		t.Errorf("%d red cards splashed, want %d", rep.Colours["R"], SplashCards)
	}
	if lands["Mountain"] != SplashLands {
		t.Errorf("%d Mountains for the splash, want %d", lands["Mountain"], SplashLands)
	}
	for _, id := range main {
		if id[0] == 'f' {
			t.Error("a card with two red symbols was splashed")
		}
	}

	// With enough cards for two full colours the deck is two colours and
	// no more, whichever two they are.
	pool = append(pool, repeatCards(fx, t, "x", 8, "C2 Fixture White Scout")...)
	pool = append(pool, repeatCards(fx, t, "y", 8, "C2 Fixture Blue Cub")...)
	main, lands = fx.rater.BuildDeck(pool, nil, 1)
	checkLegal(t, pool, nil, main, lands)
	rep = reportDeck(fx.rater, pool, main, lands)
	if len(rep.Splash) != 0 || len(rep.Main) != 2 || len(lands) != 2 || len(main) != DeckSpells {
		t.Errorf("a deck with two full colours splashed: %v, lands %v", rep.Colours, lands)
	}
}

// The edges: a pool too thin for 23 cards fills with lands and is still
// 40, an empty pool is 40 lands, and the same pool and seed give the
// same deck.
func TestBuildDeckEdgesAndDeterminism(t *testing.T) {
	fx := loadBotFixture(t)
	thin := repeatCards(fx, t, "w", 9, "C2 Fixture White Cub")
	main, lands := fx.rater.BuildDeck(thin, nil, 1)
	checkLegal(t, thin, nil, main, lands)
	if len(main) != 9 || lands["Plains"] != 31 {
		t.Errorf("a pool of nine cards: %d spells, lands %v", len(main), lands)
	}
	main, lands = fx.rater.BuildDeck(nil, nil, 1)
	checkLegal(t, nil, nil, main, lands)
	if len(main) != 0 {
		t.Errorf("an empty pool put %v in the deck", main)
	}
	// Only colourless cards: still 40, on one basic.
	rocks := repeatCards(fx, t, "j", 25, "Juggernaut")
	main, lands = fx.rater.BuildDeck(rocks, nil, 1)
	checkLegal(t, rocks, nil, main, lands)
	if len(lands) != 1 {
		t.Errorf("a colourless deck's lands are %v, want one basic", lands)
	}
	// A map that names no colour is the standard basics.
	main, lands = fx.rater.BuildDeck(thin, BasicLands{"X": "Wastes"}, 1)
	checkLegal(t, thin, nil, main, lands)

	pool := botDraft(t, fx, PickOne, 8, 9, BotPicker{Seed: 9, Rater: &fx.rater})[3]
	m1, l1 := fx.rater.BuildDeck(pool, nil, 42)
	m2, l2 := fx.rater.BuildDeck(pool, nil, 42)
	if !reflect.DeepEqual(m1, m2) || !reflect.DeepEqual(l1, l2) {
		t.Error("the same pool and seed built two decks")
	}

	// The package's BuildDeck is the same builder over the server's
	// index (UseCards).
	t.Cleanup(func() { UseCards(nil) })
	UseCards(fx.idx)
	m3, l3 := BuildDeck(pool, nil, 42)
	if !reflect.DeepEqual(m1, m3) || !reflect.DeepEqual(l1, l3) {
		t.Error("BuildDeck differs from the rater's own over the same index")
	}
	UseCards(nil)
	m4, l4 := BuildDeck(pool, nil, 42)
	checkLegal(t, pool, nil, m4, l4)
}

func TestPips(t *testing.T) {
	for cost, want := range map[string][5]float64{
		"":              {},
		"{3}":           {},
		"{1}{W}":        {1, 0, 0, 0, 0},
		"{W}{W}":        {2, 0, 0, 0, 0},
		"{X}{R}":        {0, 0, 0, 1, 0},
		"{1}{U}{R}":     {0, 1, 0, 1, 0},
		"{W/U}{G}":      {0.5, 0.5, 0, 0, 1},
		"{2/B}":         {0, 0, 1, 0, 0},
		"{G/P}":         {0, 0, 0, 0, 1},
		"{2}{B}{B}{B}":  {0, 0, 3, 0, 0},
		"{1}{R} // {U}": {0, 1, 0, 1, 0},
	} {
		if got := pips(cost); got != want {
			t.Errorf("pips(%q) = %v, want %v", cost, got, want)
		}
	}
}
