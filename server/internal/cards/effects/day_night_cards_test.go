package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// day_night_cards_test.go — day and night (CR 731, ADR 0132, #2561),
// end to end on the cards that use it. The engine rules themselves
// (the untap-step check, the daybound transform) are pinned in
// game/daynight_test.go; these tests drive them through real catalog
// entries and the real stack.

const (
	celestusOracle = "c0ad2b5f-066b-424b-bddf-d3014731e599"
)

func dnLife(p *game.Player) int { return p.Life }

// dnDesignation reads the designation without the lock (tests are
// single-goroutine).
func dnDesignation(g *game.Game) game.DayNightDesignation { return g.DayNightDesignation() }

// castPermanent casts a catalog permanent from the active seat's hand
// and settles the stack.
func dnCast(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, name, typeLine, oracle, nil)
	monSettle(t, g)
	return id
}

// The as-enters clause: neither -> day, and only from neither.
func TestCelestusBecomesDayAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	if dnDesignation(g) != game.DesignationNeither {
		t.Fatalf("a fresh game is %q", dnDesignation(g))
	}
	dnCast(t, g, "The Celestus", "Legendary Artifact", celestusOracle)
	if !g.IsDay() {
		t.Fatalf("The Celestus entered; it is %q, want day", dnDesignation(g))
	}
	// Entering with the first designation is not "day becomes night".
	if lifeFlips := countKind(g, game.EventDayNightChanged); lifeFlips != 1 {
		t.Errorf("day/night events = %d, want 1", lifeFlips)
	}
}

func TestCelestusEnteringAtNightLeavesItNight(t *testing.T) {
	g := newCatalogGame(t)
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	dnCast(t, g, "The Celestus", "Legendary Artifact", celestusOracle)
	if !g.IsNight() {
		t.Fatalf("'if it's neither day nor night' turned night into %q", dnDesignation(g))
	}
}

func countKind(g *game.Game, kind game.EventKind) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

// celestusTable is a catalog game in the active seat's main phase with
// a Celestus on the battlefield and the game already day.
func celestusTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	id := pushCatalogPermanent(g, me.ID, "The Celestus", "Legendary Artifact", celestusOracle, false)
	return g, me, id
}

func TestCelestusToggleFlipsAndTriggersItself(t *testing.T) {
	g, me, id := celestusTable(t)
	floatMana(t, g, me, "{C}{C}{C}")
	life := dnLife(me)
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the toggle: %v", err)
	}
	monSettle(t, g)
	if !g.IsNight() {
		t.Fatalf("day -> toggled: %q, want night", dnDesignation(g))
	}
	if got := dnLife(me); got != life+1 {
		t.Errorf("life = %d, want %d: the flip trigger gains 1", got, life+1)
	}
	// "You may draw a card. If you do, discard a card": yes.
	answerMayChoice(t, g, me.ID, true)
	discardFromHand(t, g, me.ID)
	monSettle(t, g)
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand = %d, want %d after a loot", got, handBefore)
	}
}

func TestCelestusLootCanBeDeclined(t *testing.T) {
	g, me, id := celestusTable(t)
	floatMana(t, g, me, "{C}{C}{C}")
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	monSettle(t, g)
	answerMayChoice(t, g, me.ID, false)
	monSettle(t, g)
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand = %d, want %d: declining the draw declines the discard", got, handBefore)
	}
	if discardOwed(g, me.ID) != 0 {
		t.Error("a declined loot left a discard prompt open")
	}
}

// Night goes back to day; the toggle is symmetric.
func TestCelestusToggleFromNightGoesToDay(t *testing.T) {
	g, me, id := celestusTable(t)
	// Night by fiat, so no trigger is left on the stack ahead of the
	// sorcery-speed activation.
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	monSettle(t, g)
	if !g.IsDay() {
		t.Fatalf("night -> toggled: %q, want day", dnDesignation(g))
	}
}

// Sorcery speed: refused on an opponent's turn, with a non-empty stack,
// and outside a main phase.
func TestCelestusToggleIsSorcerySpeed(t *testing.T) {
	g, me, id := celestusTable(t)
	floatMana(t, g, me, "{C}{C}{C}")
	// Combat is not a main phase.
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the toggle was activated outside a main phase")
	}
	if !g.IsDay() {
		t.Fatalf("a refused toggle changed the designation to %q", dnDesignation(g))
	}
}

// The trigger is the GAME's: it fires on the untap-step check of
// someone else's turn, for the Celestus's controller.
func TestCelestusTriggersOnTheUntapStepCheck(t *testing.T) {
	g, me, _ := celestusTable(t)
	life := dnLife(me)
	// Seat 0 casts nothing this turn: the next untap step makes it night.
	for g.Turn.ActiveSeat == g.StartingSeat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	monSettle(t, g)
	if !g.IsNight() {
		t.Fatalf("a quiet turn left it %q, want night", dnDesignation(g))
	}
	if got := dnLife(me); got != life+1 {
		t.Errorf("life = %d, want %d: the Celestus triggers on the turn-based flip", got, life+1)
	}
}
