package testscr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDrawAmountReplacementsAreWired(t *testing.T) {
	if n := len(game.CatalogReplacements(effectstest.ThoughtReflectionOracle)); n != 1 {
		t.Errorf("Thought Reflection declared %d replacements, want 1", n)
	}
	// Two halves: the life gain and the draw.
	if n := len(game.CatalogReplacements(effectstest.AlhammarretsArchiveOracle)); n != 2 {
		t.Errorf("Alhammarret's Archive declared %d replacements, want 2", n)
	}
}

// --- Thought Reflection ----------------------------------------------

func TestThoughtReflectionDoublesYourDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle, false)

	if got := effectstest.DrawFor(t, g, me, 1); got != 2 {
		t.Errorf("a doubled draw of one: got %d cards, want 2", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("one replacement queued %d prompts, want none", len(g.PendingChoices))
	}
}

// CR 121.2: "draw three cards" is three individual card draws, so the
// Reflection doubles each of them. Six, not four.
func TestThoughtReflectionDoublesEachOfThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle, false)

	if got := effectstest.DrawFor(t, g, me, 3); got != 6 {
		t.Errorf("a doubled \"draw three\": got %d cards, want 6", got)
	}
}

// "If YOU would draw" is the whole scope clause.
func TestThoughtReflectionLeavesAnOpponentsDrawAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle, false)

	if got := effectstest.DrawFor(t, g, opp, 2); got != 2 {
		t.Errorf("the opponent drew %d cards, want 2", got)
	}
}

// Two Thought Reflections draw FOUR, and nobody is asked to order
// them: two objects contributing ONE declared effect is #792's
// identical-window skip.
func TestTwoThoughtReflectionsDrawFourWithNoPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle, false)

	if got := effectstest.DrawFor(t, g, me, 1); got != 4 {
		t.Errorf("two Thought Reflections on one draw: got %d cards, want 4", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("two copies of one declared effect queued %d prompts, want none (#792)", len(g.PendingChoices))
	}
}

// --- Alhammarret's Archive -------------------------------------------

func TestAlhammarretsArchiveDoublesYourDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Alhammarret's Archive", "Legendary Artifact", effectstest.AlhammarretsArchiveOracle, false)
	// Off the controller's own draw step — the fixture parks the cursor
	// there, and the Archive's "except the first one you draw in each of
	// your draw steps" exempts only the first actual draw.
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })

	if got := effectstest.DrawFor(t, g, me, 1); got != 2 {
		t.Errorf("a doubled draw of one: got %d cards, want 2", got)
	}
}

// The other half, on the other event: the life gain #482 made
// replaceable.
func TestAlhammarretsArchiveDoublesYourLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.PushCatalogPermanent(g, me.ID, "Alhammarret's Archive", "Legendary Artifact", effectstest.AlhammarretsArchiveOracle, false)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.PushCatalogPermanent(g, me.ID, "Alhammarret's Archive", "Legendary Artifact", effectstest.AlhammarretsArchiveOracle, false)
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

// The fixture has already completed the turn-based draw. Later draws
// in this same draw step are doubled.
func TestAlhammarretsArchiveDoublesLaterDrawsInYourOwnDrawStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Alhammarret's Archive", "Legendary Artifact", effectstest.AlhammarretsArchiveOracle, false)
	if got := effectstest.DrawFor(t, g, me, 2); got != 4 {
		t.Errorf("later draws in the controller's draw step: got %d cards, want 4", got)
	}
}

func TestAlhammarretsArchiveExemptsOnlyTheFirstCompletedDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.PushCatalogPermanent(g, p.ID, "Alhammarret's Archive", "Legendary Artifact", effectstest.AlhammarretsArchiveOracle, false)
	before := p.Hand.Size()
	effectstest.AdvanceToDrawStepOfSeat(t, g, 1)
	if p.Hand.Size() != before+1 {
		t.Fatal("the turn-based first draw was doubled")
	}
	if got := effectstest.DrawFor(t, g, p, 1); got != 2 {
		t.Errorf("the next draw drew %d cards, want 2", got)
	}
}

// A Thought Reflection and an Archive are two DIFFERENT declared
// effects, so the affected player — the drawer — really is asked to
// order them (CR 616.1), and every ordering draws four.
func TestThoughtReflectionAndArchiveOrderAndCompose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Alhammarret's Archive", "Legendary Artifact", effectstest.AlhammarretsArchiveOracle, false)
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
