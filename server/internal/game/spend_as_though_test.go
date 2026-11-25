package game

import (
	"testing"
)

// spend_as_though_test.go — #6202: the narrower "spend mana as though"
// shapes. CR 609.4b: an effect that lets a player spend mana as though
// it were mana of another type or colour changes only how they may pay
// a cost, never the cost or the mana.
//
//	Sunglasses of Urza  white mana as though it were red mana (To)
//	Celestial Dawn      white as any colour; other mana only as colourless
//	North Star          for one spell, any type for its mana cost

const (
	whiteAsRedOracle    = "test-spend-white-as-red"
	celestialDawnOracle = "test-spend-white-any-others-colorless"
	northStarTestText   = "For one spell this turn, you may spend mana as though it were mana of any type to pay that spell's mana cost."
	whiteAsRedLabel     = "You may spend white mana as though it were red mana."
	celestialDawnLabel  = "You may spend white mana as though it were mana of any color. You may spend other mana only as though it were colorless mana."
)

func withNarrowSpendStatics(t *testing.T) {
	t.Helper()
	withAnyColorSpendStatics(t)
	prev := CatalogAnyColorSpend
	CatalogAnyColorSpend = func(key string) []AnyColorSpendStatic {
		switch key {
		case whiteAsRedOracle:
			return []AnyColorSpendStatic{{Label: whiteAsRedLabel, Whose: AnyColorSpendYou, From: "W", To: "R"}}
		case celestialDawnOracle:
			return []AnyColorSpendStatic{{Label: celestialDawnLabel, Whose: AnyColorSpendYou, From: "W", OthersOnlyColorless: true}}
		}
		return prev(key)
	}
	t.Cleanup(func() { CatalogAnyColorSpend = prev })
}

func narrowSpendTable(t *testing.T) (*Game, *Player, *Player) {
	t.Helper()
	withNarrowSpendStatics(t)
	g := newActiveGame(t)
	toMainPhase(t, g)
	return g, g.Seats[0], g.Seats[1]
}

// Sunglasses of Urza: white pays a {R} (and the red half of a hybrid),
// never a {U}.
func TestSpendWhiteAsRedPaysRedOnly(t *testing.T) {
	g, me, _ := narrowSpendTable(t)
	pushSpendGrant(g, me, whiteAsRedOracle)
	red := pushHandSpell(me, "Red Spell", "Instant", "{R}")
	hybrid := pushHandSpell(me, "Hybrid Spell", "Instant", "{R/G}")
	blue := pushHandSpell(me, "Blue Spell", "Instant", "{U}")

	me.ManaPool.AddMana(ManaToken{Color: "W"})
	wantInsufficient(t, g.CastSpell(me.ID, blue, CastSpellParams{Strict: true}), "white for {U} under white-as-red")
	if err := g.CastSpell(me.ID, red, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("white for {R}: %v", err)
	}
	me.ManaPool.AddMana(ManaToken{Color: "W"})
	if err := g.CastSpell(me.ID, hybrid, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("white for {R/G}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}

// The grant is white-as-red, not red-as-white, and it is the
// controller's only.
func TestSpendWhiteAsRedIsOneWayAndTheControllers(t *testing.T) {
	g, me, opp := narrowSpendTable(t)
	pushSpendGrant(g, me, whiteAsRedOracle)
	white := pushHandSpell(me, "White Spell", "Instant", "{W}")
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	wantInsufficient(t, g.CastSpell(me.ID, white, CastSpellParams{Strict: true}), "red for {W} under white-as-red")

	theirs := pushHandSpell(opp, "Their Red Spell", "Instant", "{R}")
	opp.ManaPool.AddMana(ManaToken{Color: "W"})
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 })
	wantInsufficient(t, g.CastSpell(opp.ID, theirs, CastSpellParams{Strict: true}), "white for an opponent's {R}")
}

// The auto-tapper plans the cost the payment charges: a Plains pays a {R}.
func TestSpendWhiteAsRedThroughTheAutoTapper(t *testing.T) {
	g, me, _ := narrowSpendTable(t)
	pushSpendGrant(g, me, whiteAsRedOracle)
	pushUntappedLands(t, g, me, 1, "Basic Land — Plains")
	spell := pushHandSpell(me, "Red Spell", "Instant", "{R}")
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {R} off a Plains: %v", err)
	}
}

// Celestial Dawn: white pays any coloured symbol; red mana is spent only
// as though it were colourless, so it pays generic and a {C}, never a
// {R} — even a red symbol.
func TestCelestialDawnOtherManaOnlyAsColorless(t *testing.T) {
	g, me, _ := narrowSpendTable(t)
	pushSpendGrant(g, me, celestialDawnOracle)
	red := pushHandSpell(me, "Red Spell", "Instant", "{R}")
	generic := pushHandSpell(me, "Generic Spell", "Instant", "{1}")
	seer := pushHandSpell(me, "Colorless Spell", "Instant", "{C}")
	blue := pushHandSpell(me, "Blue Spell", "Instant", "{U}")

	me.ManaPool.AddMana(ManaToken{Color: "R"})
	wantInsufficient(t, g.CastSpell(me.ID, red, CastSpellParams{Strict: true}), "red for {R} under Celestial Dawn")
	if err := g.CastSpell(me.ID, generic, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("red for {1}: %v", err)
	}
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, seer, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("red as colorless for {C}: %v", err)
	}
	me.ManaPool.AddMana(ManaToken{Color: "W"})
	if err := g.CastSpell(me.ID, blue, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("white as any colour for {U}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}

// White is not "other mana": it is spent as white or as any colour, and
// colourless is not a colour (CR 106.1b), so it still can't pay a {C}.
func TestCelestialDawnWhiteDoesNotPayColorless(t *testing.T) {
	g, me, _ := narrowSpendTable(t)
	pushSpendGrant(g, me, celestialDawnOracle)
	seer := pushHandSpell(me, "Colorless Spell", "Instant", "{C}")
	me.ManaPool.AddMana(ManaToken{Color: "W"})
	wantInsufficient(t, g.CastSpell(me.ID, seer, CastSpellParams{Strict: true}), "white for {C} under Celestial Dawn")
}

// The restriction is its controller's alone: an opponent's red still
// pays their {R}.
func TestCelestialDawnBindsOnlyItsController(t *testing.T) {
	g, me, opp := narrowSpendTable(t)
	pushSpendGrant(g, me, celestialDawnOracle)
	theirs := pushHandSpell(opp, "Their Red Spell", "Instant", "{R}")
	opp.ManaPool.AddMana(ManaToken{Color: "R"})
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 })
	if err := g.CastSpell(opp.ID, theirs, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("opponent's red for their {R}: %v", err)
	}
}

func TestRestrictToColorlessShapes(t *testing.T) {
	cost, err := ParseCost("{2}{R}{C}{S}")
	if err != nil {
		t.Fatal(err)
	}
	got := restrictToColorless(cost, []string{"W"})
	if got.Generic != 2 {
		t.Errorf("generic = %d, want 2", got.Generic)
	}
	for _, r := range got.Required {
		switch {
		case r.Snow:
			if r.Admits("W") {
				t.Errorf("snow slot widened: %+v", r)
			}
		case r.Admits("C"):
			if r.Admits("W") || !r.Admits("R") || !r.Admits("G") {
				t.Errorf("{C} slot = %v, want C and every non-white colour", r.Options)
			}
		default:
			if !r.Admits("W") || r.Admits("R") {
				t.Errorf("{R} slot = %v, want white only", r.Options)
			}
		}
	}
}

// grantNorthStar gives `p` North Star's one-spell any-type promise.
func grantNorthStar(g *Game, p *Player) {
	g.WithWriteLock(func() {
		g.GrantNextSpellPromiseForEffect(p.ID, NextSpellPromise{ManaCostAnyType: true, Text: northStarTestText}, "North Star", p.ID, Duration{})
	})
}

func liveNorthStars(g *Game, p *Player) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, s := range p.Statics {
			if s.NextSpell.ManaCostAnyType && g.livePromise(s) {
				n++
			}
		}
	})
	return n
}

// North Star: one spell's {C}{U} is paid with red mana, and the second
// spell pays normally — the promise is spent.
func TestNorthStarPaysOneSpellsManaCostWithAnyType(t *testing.T) {
	g, me, _ := narrowSpendTable(t)
	grantNorthStar(g, me)
	first := pushHandSpell(me, "First", "Instant", "{C}{U}")
	second := pushHandSpell(me, "Second", "Instant", "{U}")
	me.ManaPool.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, first, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("{C}{U} off two red under North Star: %v", err)
	}
	if n := liveNorthStars(g, me); n != 0 {
		t.Fatalf("%d North Star promises live after the spell, want 0", n)
	}
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	wantInsufficient(t, g.CastSpell(me.ID, second, CastSpellParams{Strict: true}), "red for {U} after North Star was spent")
}

// A generic-only spell is one North Star cannot matter to, so it is
// left for the spell it was activated for.
func TestNorthStarIsNotSpentOnAGenericSpell(t *testing.T) {
	g, me, _ := narrowSpendTable(t)
	grantNorthStar(g, me)
	generic := pushHandSpell(me, "Generic", "Instant", "{1}")
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, generic, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast {1}: %v", err)
	}
	if n := liveNorthStars(g, me); n != 1 {
		t.Fatalf("%d North Star promises live, want 1", n)
	}
}

// "This turn": the promise lapses at end of turn.
func TestNorthStarLapsesAtEndOfTurn(t *testing.T) {
	g, me, _ := narrowSpendTable(t)
	grantNorthStar(g, me)
	endTheTurn(g)
	if n := liveNorthStars(g, me); n != 0 {
		t.Fatalf("%d North Star promises live next turn, want 0", n)
	}
}

// The mana COST only: the widening is applied before additional costs
// join the price, so it reaches the printed symbols and nothing else.
func TestNorthStarWidensOnlyThePrintedManaCost(t *testing.T) {
	g, me, _ := narrowSpendTable(t)
	grantNorthStar(g, me)
	card := NewCard("Spell", me.ID)
	card.TypeLine = "Instant"
	card.ManaCost = "{U}"
	var cost ParsedCost
	g.ReadSnapshot(func() {
		var err error
		cost, _, err = g.printedCostLocked(me, card, CastSpellParams{})
		if err != nil {
			t.Fatal(err)
		}
	})
	if len(cost.Required) != 1 || !cost.Required[0].AnyMana {
		t.Fatalf("priced %+v, want the {U} widened to any mana", cost.Required)
	}
}
