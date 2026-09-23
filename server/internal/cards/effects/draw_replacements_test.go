package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draw_replacements_test.go — #1222 from the card side: the CR 614
// window on the draw AMOUNT, watched through the ordinary
// Spec.Replacements slot.
//
// Two printed cards again: Thought Reflection's bare clause and
// Alhammarret's Archive's same clause with "except the first one you
// draw in each of your draw steps" (read as "except any draw in your
// own draw step" — Notion Thief's declared simplification) plus a life
// half on a different event.

const (
	thoughtReflectionOracle   = "692a6833-3014-42f6-b1ad-333bb3292c65"
	alhammarretsArchiveOracle = "d5c4d36c-54b3-4149-906b-a57a670260fc"
)

// drawFor runs an effect-side draw of n cards for `seat` under the
// write lock and reports how many cards reached the hand.
func drawFor(t *testing.T, g *game.Game, seat *game.Player, n int) int {
	t.Helper()
	before := seat.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(seat.ID, n); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	return seat.Hand.Size() - before
}

func TestDrawAmountReplacementsAreWired(t *testing.T) {
	if n := len(game.CatalogReplacements(thoughtReflectionOracle)); n != 1 {
		t.Errorf("Thought Reflection declared %d replacements, want 1", n)
	}
	// Two halves: the life gain and the draw.
	if n := len(game.CatalogReplacements(alhammarretsArchiveOracle)); n != 2 {
		t.Errorf("Alhammarret's Archive declared %d replacements, want 2", n)
	}
}

// --- Thought Reflection ----------------------------------------------

func TestThoughtReflectionDoublesYourDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", thoughtReflectionOracle, false)

	if got := drawFor(t, g, me, 1); got != 2 {
		t.Errorf("a doubled draw of one: got %d cards, want 2", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("one replacement queued %d prompts, want none", len(g.PendingChoices))
	}
}

// CR 121.2: "draw three cards" is three individual card draws, so the
// Reflection doubles each of them. Six, not four.
func TestThoughtReflectionDoublesEachOfThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", thoughtReflectionOracle, false)

	if got := drawFor(t, g, me, 3); got != 6 {
		t.Errorf("a doubled \"draw three\": got %d cards, want 6", got)
	}
}

// "If YOU would draw" is the whole scope clause.
func TestThoughtReflectionLeavesAnOpponentsDrawAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", thoughtReflectionOracle, false)

	if got := drawFor(t, g, opp, 2); got != 2 {
		t.Errorf("the opponent drew %d cards, want 2", got)
	}
}

// Two Thought Reflections draw FOUR, and nobody is asked to order
// them: two objects contributing ONE declared effect is #792's
// identical-window skip.
func TestTwoThoughtReflectionsDrawFourWithNoPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", thoughtReflectionOracle, false)
	pushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", thoughtReflectionOracle, false)

	if got := drawFor(t, g, me, 1); got != 4 {
		t.Errorf("two Thought Reflections on one draw: got %d cards, want 4", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("two copies of one declared effect queued %d prompts, want none (#792)", len(g.PendingChoices))
	}
}

// --- Alhammarret's Archive -------------------------------------------

func TestAlhammarretsArchiveDoublesYourDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Alhammarret's Archive", "Legendary Artifact", alhammarretsArchiveOracle, false)
	// Off the controller's own draw step — the fixture parks the cursor
	// there, and the Archive's "except the first one you draw in each of
	// your draw steps" declines the whole step (the simplification, and
	// its own test below).
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })

	if got := drawFor(t, g, me, 1); got != 2 {
		t.Errorf("a doubled draw of one: got %d cards, want 2", got)
	}
}

// The other half, on the other event: the life gain #482 made
// replaceable.
func TestAlhammarretsArchiveDoublesYourLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushCatalogPermanent(g, me.ID, "Alhammarret's Archive", "Legendary Artifact", alhammarretsArchiveOracle, false)
	before := me.Life

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(src, me.ID, 3); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
	if got := me.Life - before; got != 6 {
		t.Errorf("a doubled gain of 3: got %d life, want 6", got)
	}
}

// A life LOSS is not a gain, and goes through untouched.
func TestAlhammarretsArchiveLeavesALifeLossAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushCatalogPermanent(g, me.ID, "Alhammarret's Archive", "Legendary Artifact", alhammarretsArchiveOracle, false)
	before := me.Life

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(src, me.ID, -3); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
	if got := me.Life - before; got != -3 {
		t.Errorf("a loss of 3: got %d, want -3", got)
	}
}

// THE DECLARED SIMPLIFICATION: "except the first one you draw in each
// of your draw steps" is read as "except any draw in your own draw
// step". newCatalogGame parks the cursor on seat 0's draw step, so a
// draw for seat 0 there is left alone and the same draw for seat 1 is
// doubled.
func TestAlhammarretsArchiveSkipsYourOwnDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Alhammarret's Archive", "Legendary Artifact", alhammarretsArchiveOracle, false)
	if g.Turn.Step != game.StepDraw || g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Fatalf("fixture: expected the cursor on seat 0's draw step, got %v seat %d", g.Turn.Step, g.Turn.ActiveSeat)
	}

	if got := drawFor(t, g, me, 1); got != 1 {
		t.Errorf("a draw in the controller's own draw step: got %d cards, want 1 (the simplification)", got)
	}
}

// A Thought Reflection and an Archive are two DIFFERENT declared
// effects, so the affected player — the drawer — really is asked to
// order them (CR 616.1), and every ordering draws four.
func TestThoughtReflectionAndArchiveOrderAndCompose(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", thoughtReflectionOracle, false)
	pushCatalogPermanent(g, me.ID, "Alhammarret's Archive", "Legendary Artifact", alhammarretsArchiveOracle, false)
	// Off the controller's own draw step, or the Archive's clause
	// declines and there is only one applicable replacement.
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })

	before := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("two different draw replacements is a CR 616 ordering prompt: %+v", g.PendingChoices)
	}
	c := g.PendingChoices[0]
	if c.Chooser != me.ID {
		t.Errorf("CR 616.1 asks the affected player (the drawer): got %v, want %v", c.Chooser, me.ID)
	}
	if err := g.ResolveReplacementOrder(c.ID, c.Chooser, c.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if got := me.Hand.Size() - before; got != 4 {
		t.Errorf("Thought Reflection + Alhammarret's Archive on one draw: got %d cards, want 4", got)
	}
}
