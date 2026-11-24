package game

import (
	"testing"

	"github.com/google/uuid"
)

// draw_instruction_rewrite_test.go — #6089: Alms Collector's "If an
// opponent would draw two or more cards, instead you and that player each
// draw a card" (draw_instruction_rewrite.go).

func almsForTest(controller uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches:                    []EventKind{EventDrawCard},
		DrawInstructionEachDrawOne: true,
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventDraw && ev.DrawInstruction && ev.DrawCount >= 2 && ev.DrawPlayer != controller
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return controller },
		Label:      "alms",
	}
}

func riddlerForTest(owner uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches:              []EventKind{EventDrawCard},
		DrawInstructionExtra: 1,
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventDraw && ev.DrawInstruction && ev.DrawCount >= 1 && ev.DrawPlayer == owner
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      "riddler",
	}
}

// "Draw three" by an opponent is one card for them and one for the
// collector; "draw one" is untouched; the collector's own draws are
// untouched.
func TestDrawInstructionRewriteEachDrawOne(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	stockLibrary(me, 10)
	stockLibrary(opp, 10)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(almsForTest(me.ID)) })

	mine, theirs := handSize(me), handSize(opp)
	from := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(opp.ID, 3); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	if got := handSize(opp) - theirs; got != 1 {
		t.Errorf("the opponent drew %d, want 1", got)
	}
	if got := handSize(me) - mine; got != 1 {
		t.Errorf("the collector drew %d, want 1", got)
	}
	if countDrawEventsSince(g, from, me.ID) != 1 || countDrawEventsSince(g, from, opp.ID) != 1 {
		t.Error("each draw should be one EventDrawCard by its own drawer")
	}

	mine, theirs = handSize(me), handSize(opp)
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(opp.ID, 1); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
		if err := g.DrawNForEffect(me.ID, 2); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	if got := handSize(opp) - theirs; got != 1 {
		t.Errorf("a one-card draw was rewritten: opponent drew %d, want 1", got)
	}
	if got := handSize(me) - mine; got != 2 {
		t.Errorf("the collector's own draw two was rewritten: drew %d, want 2", got)
	}
}

// The "then" of the instruction runs once, after both draws.
func TestDrawInstructionRewriteRunsThenAfterBoth(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	stockLibrary(me, 10)
	stockLibrary(opp, 10)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(almsForTest(me.ID)) })
	var sawMine, sawTheirs, runs int
	ref := RegisterDrawThen(testEffectKeyPrefix+"alms-then", func(g *Game, d DrawThen) error {
		runs++
		sawMine, sawTheirs = g.Seats[0].Hand.Size(), g.Seats[1].Hand.Size()
		return nil
	})
	mine, theirs := handSize(me), handSize(opp)
	g.WithWriteLock(func() {
		if err := g.DrawNThenForEffect(opp.ID, 2, DrawThen{Ref: ref, Player: opp.ID}); err != nil {
			t.Fatalf("DrawNThenForEffect: %v", err)
		}
	})
	if runs != 1 || sawMine != mine+1 || sawTheirs != theirs+1 {
		t.Errorf("then ran %d times seeing hands %d/%d, want once after both draws (%d/%d)", runs, sawMine, sawTheirs, mine+1, theirs+1)
	}
}

// With the drawer's own Riddler beside it, the rewrite applies first and
// the Riddler adds its one to the drawer's one.
func TestDrawInstructionRewriteBeforeExtra(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	stockLibrary(me, 10)
	stockLibrary(opp, 10)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(almsForTest(me.ID))
		g.RegisterReplacementForTest(riddlerForTest(opp.ID))
	})
	mine, theirs := handSize(me), handSize(opp)
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(opp.ID, 2); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	if got := handSize(opp) - theirs; got != 2 {
		t.Errorf("the opponent drew %d, want 2 (one, plus the Riddler's one)", got)
	}
	if got := handSize(me) - mine; got != 1 {
		t.Errorf("the collector drew %d, want 1", got)
	}
}

// draw_from_bottom (#6089): River Song's controller draws the bottom card;
// everybody else, and the controller once the permanent loses its key, the
// top one.
func TestDrawsFromBottom(t *testing.T) {
	prev := CatalogDrawsFromBottom
	CatalogDrawsFromBottom = func(key string) bool { return key == "test-river" }
	t.Cleanup(func() { CatalogDrawsFromBottom = prev })
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	mine := searchLibraryOf(me, "bottom", "middle", "top")
	theirs := searchLibraryOf(opp, "bottom", "middle", "top")
	river := pushBattlefieldForTest(g, me.ID, "River", "Legendary Creature — Human", "test-river")
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.DrawNForEffect(opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	if !me.Hand.Contains(mine[0]) || me.Hand.Contains(mine[2]) {
		t.Error("the controller did not draw the bottom card")
	}
	if !opp.Hand.Contains(theirs[2]) || opp.Hand.Contains(theirs[0]) {
		t.Error("an opponent did not draw the top card")
	}
	g.WithWriteLock(func() {
		if _, err := g.Battlefield.Remove(river); err != nil {
			t.Fatal(err)
		}
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	if !me.Hand.Contains(mine[2]) {
		t.Error("with the permanent gone the controller did not draw the top card")
	}
}
