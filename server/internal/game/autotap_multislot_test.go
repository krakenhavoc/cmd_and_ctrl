package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// autotap_multislot_test.go is #2461: a source whose ONE activation
// makes two or more mana of different colours (a bounce land's
// "{T}: Add {G}{U}") is booked slot by slot, never one slot twice, and
// a cast the plan cannot fund is refused with nothing tapped and
// nothing floating.

// multiSlotHook serves the made-up sources these tests use.
func multiSlotHook(oracleID string) []ManaAbilityShape {
	switch oracleID {
	case "zz-chamber": // Simic Growth Chamber
		return []ManaAbilityShape{{TapCost: true, Produced: "{G}{U}", Label: "Add {G}{U}"}}
	case "zz-boilerworks": // Izzet Boilerworks
		return []ManaAbilityShape{{TapCost: true, Produced: "{U}{R}", Label: "Add {U}{R}"}}
	case "zz-island":
		return []ManaAbilityShape{{TapCost: true, Produced: "{U}", Label: "Add {U}"}}
	case "zz-mountain":
		return []ManaAbilityShape{{TapCost: true, Produced: "{R}", Label: "Add {R}"}}
	case "zz-tomb": // Ancient Tomb's shape, without the pain
		return []ManaAbilityShape{{TapCost: true, Produced: "{C}{C}", Label: "Add {C}{C}"}}
	case "zz-split": // a hybrid slot beside a plain one
		return []ManaAbilityShape{{TapCost: true, Produced: "{W|U}{U}", Label: "Add {W|U}{U}"}}
	}
	return nil
}

func tappedCount(g *Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Tapped {
			n++
		}
	}
	return n
}

// TestBounceLandIsNotBookedTwice is the issue's reproduction: one
// Simic Growth Chamber is not {U}{U}, the planner says so, and the
// strict auto-tapped Counterspell is refused with the land untapped
// and the pool empty.
func TestBounceLandIsNotBookedTwice(t *testing.T) {
	withCatalogHook(t, multiSlotHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Simic Growth Chamber", "Land", "zz-chamber")
	id := pushTypedCardToHandWithCost(p, "Counterspell", "Instant", "{U}{U}")

	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}{U}"), 0); ok {
		t.Fatal("planner: one {G}{U} land was planned as paying {U}{U}")
	}
	err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true})
	var ime *InsufficientManaError
	if !errors.As(err, &ime) {
		t.Fatalf("cast: want InsufficientManaError, got %v", err)
	}
	if n := tappedCount(g, p.ID); n != 0 {
		t.Errorf("refused cast left %d permanent(s) tapped", n)
	}
	if got := len(g.Seats[0].ManaPool); got != 0 {
		t.Errorf("refused cast left %d mana floating: %v", got, g.Seats[0].ManaPool)
	}
}

// TestBounceLandPaysBothOfItsColours: the chamber's two slots are two
// different pips, in either order, and its second slot still pays
// generic once one colour is booked.
func TestBounceLandPaysBothOfItsColours(t *testing.T) {
	withCatalogHook(t, multiSlotHook)
	for _, cost := range []string{"{G}{U}", "{U}{G}", "{1}{U}", "{1}{G}"} {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		p := g.Seats[0]
		pushBattlefieldForTest(g, p.ID, "Simic Growth Chamber", "Land", "zz-chamber")
		if _, ok := g.AutoTapForCost(p.ID, costFor(t, cost), 0); !ok {
			t.Errorf("%s: one {G}{U} land should pay it", cost)
		}
		id := pushTypedCardToHandWithCost(p, "Spell "+cost, "Instant", cost)
		if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
			t.Errorf("%s: cast refused: %v", cost, err)
		}
	}
}

// TestBoilerworksDoesNotPayBothRedPips is the Glint-Horn Buccaneer
// game: Izzet Boilerworks and an Island are {U}{R}{U}, which is not
// {1}{R}{R}. A Mountain beside them makes it payable.
func TestBoilerworksDoesNotPayBothRedPips(t *testing.T) {
	withCatalogHook(t, multiSlotHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Izzet Boilerworks", "Land", "zz-boilerworks")
	pushBattlefieldForTest(g, p.ID, "Island", "Land", "zz-island")
	id := pushTypedCardToHandWithCost(p, "Glint-Horn Buccaneer", "Creature — Minotaur Pirate", "{1}{R}{R}")

	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{1}{R}{R}"), 0); ok {
		t.Fatal("planner: Boilerworks + Island were planned as paying {1}{R}{R}")
	}
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("cast should be refused")
	}
	if n := tappedCount(g, p.ID); n != 0 {
		t.Errorf("refused cast left %d permanent(s) tapped", n)
	}
	if got := len(g.Seats[0].ManaPool); got != 0 {
		t.Errorf("refused cast left mana floating: %v", g.Seats[0].ManaPool)
	}

	pushBattlefieldForTest(g, g.Seats[0].ID, "Mountain", "Land", "zz-mountain")
	if err := g.CastSpell(g.Seats[0].ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("with a Mountain: %v", err)
	}
}

// TestUniformMultiSlotSourcesStillPay: Sol Ring / Ancient Tomb's two
// identical slots pay two generic, and two {C} pips.
func TestUniformMultiSlotSourcesStillPay(t *testing.T) {
	withCatalogHook(t, multiSlotHook)
	for _, cost := range []string{"{2}", "{C}{C}", "{C}{1}"} {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		p := g.Seats[0]
		pushBattlefieldForTest(g, p.ID, "Ancient Tomb", "Land", "zz-tomb")
		if _, ok := g.AutoTapForCost(p.ID, costFor(t, cost), 0); !ok {
			t.Errorf("%s: a {C}{C} source should pay it", cost)
		}
	}
	g := newActiveGame(t)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Ancient Tomb", "Land", "zz-tomb")
	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{3}"), 0); ok {
		t.Error("{3}: a {C}{C} source alone should not pay it")
	}
}

// TestSlotChoiceBacktracks: a {W|U}{U} source pays {U}{W} only if the
// {U} pip takes the plain {U} slot and leaves the hybrid one for {W}.
// The solver has to be able to try the second slot, not only the first
// free one.
func TestSlotChoiceBacktracks(t *testing.T) {
	withCatalogHook(t, multiSlotHook)
	g := newActiveGame(t)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Split Rock", "Artifact", "zz-split")
	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{U}{W}"), 0); !ok {
		t.Error("{U}{W}: the {W|U}{U} source pays it with the {U} pip on its second slot")
	}
	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{W}{W}"), 0); ok {
		t.Error("{W}{W}: the {W|U}{U} source makes one {W} at most")
	}
}

// TestRefusedAutoTapCastChangesNothing is the all-or-nothing half,
// with the planner and the executor made to disagree on purpose: a
// source whose gate holds while the plan is made and stops holding
// once anything has been tapped. The plan books it; the executor drops
// it. The cast must be refused with the Forest it would have tapped
// first still untapped and nothing floating.
func TestRefusedAutoTapCastChangesNothing(t *testing.T) {
	withCatalogHook(t, func(o string) []ManaAbilityShape {
		switch o {
		case "zz-forest":
			return []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}
		case "zz-fickle":
			return []ManaAbilityShape{{
				TapCost: true, Produced: "{U}", Label: "Add {U}",
				// Holds only while no permanent of the controller's is
				// tapped: true when the plan is made, false once the
				// executor has tapped the Forest.
				Condition: func(g *Game, controller, _ uuid.UUID) bool {
					return tappedCount(g, controller) == 0
				},
			}}
		}
		return nil
	})
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Forest", "Land", "zz-forest")
	pushBattlefieldForTest(g, p.ID, "Fickle Spring", "Land", "zz-fickle")
	id := pushTypedCardToHandWithCost(p, "Two Colours", "Instant", "{G}{U}")

	if _, ok := g.AutoTapForCost(p.ID, costFor(t, "{G}{U}"), 0); !ok {
		t.Fatal("setup: the planner should book both lands")
	}
	events := len(g.Events)
	err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true})
	var ime *InsufficientManaError
	if !errors.As(err, &ime) {
		t.Fatalf("cast: want InsufficientManaError, got %v", err)
	}
	if n := tappedCount(g, g.Seats[0].ID); n != 0 {
		t.Errorf("refused cast left %d permanent(s) tapped", n)
	}
	if got := len(g.Seats[0].ManaPool); got != 0 {
		t.Errorf("refused cast left mana floating: %v", g.Seats[0].ManaPool)
	}
	if got := len(g.Events); got != events {
		t.Errorf("refused cast emitted %d event(s): %v", got-events, g.Events[events:])
	}
	if !g.Seats[0].Hand.Contains(id) {
		t.Error("refused cast moved the card out of the hand")
	}
}
