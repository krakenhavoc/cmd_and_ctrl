package game

import (
	"testing"

	"github.com/google/uuid"
)

// draw_count_test.go — #1222, the engine half of the draw AMOUNT. The
// catalog half (the real Thought Reflection and Alhammarret's Archive)
// is in cards/effects/draw_replacements_test.go.
//
// The draw event has existed since S17 and Notion Thief has been
// rewriting its DRAWING PLAYER since S20. What there was no room on it
// for is HOW MANY, and what is pinned here is the three things the
// count has to be true of: it is one per instruction (CR 121.2), the
// cards it settles on are drawn one at a time, and the whole family of
// "instead" replacements still composes through CR 616.1.

// drawCountReplacement is "if <owner> would draw a card, they draw
// count(n) instead" as a test injection.
func drawCountReplacement(owner uuid.UUID, count func(int) int, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventDrawCard},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventDraw && ev.DrawCount > 0 && ev.DrawPlayer == owner
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.DrawCount = count(ev.DrawCount)
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      label,
	}
}

// dredgeStyleReplacement stands in for dredge: an "instead" that
// cancels the draw and mills instead. No dredge card is catalogued, so
// this is the shape rather than the card — what it pins is that a
// cancel-style draw replacement and a count-style one share one window
// and compose through the apply-loop.
func dredgeStyleReplacement(owner uuid.UUID, milled *int, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventDrawCard},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventDraw && ev.DrawPlayer == owner
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			*milled += ev.DrawCount
			ev.Cancel()
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      label,
	}
}

func handSize(p *Player) int { return p.Hand.Size() }

// countDrawEventsSince tallies EventDrawCard for one player from
// `from` on.
func countDrawEventsSince(g *Game, from int, player uuid.UUID) int {
	n := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventDrawCard && ev.Actor == player {
			n++
		}
	}
	return n
}

// TestDrawCountDoublesOneDraw is the base case: one instruction, one
// window, two cards.
func TestDrawCountDoublesOneDraw(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	stockLibrary(p, 10)
	before := handSize(p)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(drawCountReplacement(p.ID, timesTwoForTest, "double"))
		if err := g.DrawNForEffect(p.ID, 1); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	if got := handSize(p) - before; got != 2 {
		t.Errorf("a doubled draw of one: got %d cards, want 2", got)
	}
}

// TestDrawCountIsPerInstruction is CR 121.2: "draw three cards" is
// three individual card draws, so each opens its own window and the
// doubler doubles each. Six, not four.
func TestDrawCountIsPerInstruction(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	stockLibrary(p, 20)
	before := handSize(p)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(drawCountReplacement(p.ID, timesTwoForTest, "double"))
		if err := g.DrawNForEffect(p.ID, 3); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	if got := handSize(p) - before; got != 6 {
		t.Errorf("a doubled \"draw three\": got %d cards, want 6", got)
	}
}

// TestDrawCountEmitsOneEventPerCard is why the N cards are drawn one
// at a time: every per-card payoff in the catalog reads EventDrawCard.
func TestDrawCountEmitsOneEventPerCard(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	stockLibrary(p, 10)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(drawCountReplacement(p.ID, timesTwoForTest, "double"))
	})
	from := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(p.ID, 1); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	if got := countDrawEventsSince(g, from, p.ID); got != 2 {
		t.Errorf("EventDrawCard for a doubled draw: got %d, want 2 (one per card)", got)
	}
}

// TestTwoDrawDoublersCompose is the printed ruling: two Thought
// Reflections draw FOUR, not three. The second multiplies what the
// first left, through the ordinary CR 616.1 apply-loop, and CR 614.5's
// once-per-event tracking is what stops either from applying twice.
func TestTwoDrawDoublersCompose(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	stockLibrary(p, 20)
	before := handSize(p)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(drawCountReplacement(p.ID, timesTwoForTest, "double A"))
		g.RegisterReplacementForTest(drawCountReplacement(p.ID, timesTwoForTest, "double B"))
	})
	// Two DIFFERENT declared effects is a CR 616 ordering prompt; the
	// engine queues it and the resume finishes the draw. Answer it in
	// the offered order.
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(p.ID, 1); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) == 1 && g.PendingChoices[0].Kind == PendingChoiceReplacementOrder {
		c := g.PendingChoices[0]
		if err := g.ResolveReplacementOrder(c.ID, c.Chooser, c.ReplacementEffectIDs); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
	}
	if got := handSize(p) - before; got != 4 {
		t.Errorf("two draw doublers on one draw: got %d cards, want 4", got)
	}
}

// TestDredgeStyleReplacementComposesWithADoubler is the "instead"
// family sharing one window. Dredge first and the whole draw is
// replaced; the doubler first and the dredge takes the doubled draw.
// Either way the affected player — the DRAWER — is the one CR 616.1
// asks (#982), and either way the pipeline reaches a terminal answer.
func TestDredgeStyleReplacementComposesWithADoubler(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	stockLibrary(p, 20)
	before := handSize(p)
	dredged := 0
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(drawCountReplacement(p.ID, timesTwoForTest, "double"))
		g.RegisterReplacementForTest(dredgeStyleReplacement(p.ID, &dredged, "dredge"))
		if err := g.DrawNForEffect(p.ID, 1); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceReplacementOrder {
		t.Fatalf("two different draw replacements is a CR 616 ordering prompt: %+v", g.PendingChoices)
	}
	c := g.PendingChoices[0]
	if c.Chooser != p.ID {
		t.Errorf("CR 616.1 asks the AFFECTED player (the drawer): got %v, want %v", c.Chooser, p.ID)
	}
	// Doubler first: the dredge then sees a draw of two and takes it.
	order := c.ReplacementEffectIDs
	if err := g.ResolveReplacementOrder(c.ID, c.Chooser, order); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if handSize(p) != before {
		t.Errorf("a dredged draw draws nothing: hand %d, want %d", handSize(p), before)
	}
	if dredged == 0 {
		t.Errorf("the dredge-style replacement never fired")
	}
}

// TestDrawCountStopsAtAnEmptyLibrary keeps DrawCard's wire behaviour
// and DrawNForEffect's partial-draw contract: the second card of a
// doubled draw against a one-card library sets the CR 704.5b flag.
func TestDrawCountStopsAtAnEmptyLibrary(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	stockLibrary(p, 1)
	before := handSize(p)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(drawCountReplacement(p.ID, timesTwoForTest, "double"))
		if err := g.DrawNForEffect(p.ID, 1); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	if got := handSize(p) - before; got != 1 {
		t.Errorf("a doubled draw against a one-card library: got %d, want 1", got)
	}
	if !p.AttemptedEmptyDraw {
		t.Errorf("the second card of the doubled draw came off an empty library (CR 704.5b)")
	}
}
