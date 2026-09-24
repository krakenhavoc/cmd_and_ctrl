package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_batch_clone_test.go — #1322's copy half. The copy choice needs
// a resumable entry (copy_choice.go), and the put batch was not one, so
// a Clone put by Arboreal Grazer or Genesis Wave entered as the 0/0 it
// prints and died.

// ebCloneInHand puts a Clone into seat 0's hand beside a Grizzly Bears
// on the battlefield.
func ebCloneInHand(g *game.Game) (me *game.Player, clone, bears uuid.UUID) {
	me = g.Seats[0]
	bears = seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	clone = uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: clone, Name: "Clone", TypeLine: "Creature — Shapeshifter",
		OracleID: oracleClone, Owner: me.ID, Controller: me.ID,
	})
	return me, clone, bears
}

// TestCloneThatASpellPutsOntoTheBattlefieldCopies: it is asked now, and
// enters as the copy, and the put's continuation is told once it has.
func TestCloneThatASpellPutsOntoTheBattlefieldCopies(t *testing.T) {
	g := newCatalogGame(t)
	me, clone, bears := ebCloneInHand(g)
	var entered uuid.UUID
	g.WithWriteLock(func() {
		if err := g.PutFromHandOntoBattlefieldThenForEffect(clone, game.HandEntryOptions{Controller: me.ID},
			func(_ *game.Game, id uuid.UUID) error { entered = id; return nil }); err != nil {
			t.Fatalf("put: %v", err)
		}
	})
	c := copyPrompt(g)
	if c == nil {
		t.Fatal("a Clone put onto the battlefield by a spell was not asked what to copy")
	}
	if entered != uuid.Nil {
		t.Fatal("the continuation ran while the copy question was open")
	}
	if err := g.ResolveCopyTarget(c.ID, c.Chooser, bears); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	if entered != clone {
		t.Fatalf("continuation told %s, want the Clone", entered)
	}
	if got := copyBattlefieldCard(t, g, clone); got.Name != "Grizzly Bears" {
		t.Errorf("the Clone entered as %q, want a copy of Grizzly Bears", got.Name)
	}
}

// TestACloneEntryCancelledAfterItsCopyStillFinishesTheBatch: the copy
// answer is given, and then another replacement cancels the entry
// outright. The copy prompt's resume used to return nil for a cancelled
// entry, so the batch it belonged to — and the continuation behind it —
// waited forever. It goes through the shared finisher now, which tells
// the batch nothing entered.
func TestACloneEntryCancelledAfterItsCopyStillFinishesTheBatch(t *testing.T) {
	g := newCatalogGame(t)
	me, clone, bears := ebCloneInHand(g)
	thenRan := 0
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.CardID == clone &&
					ev.NewZone == game.ZoneBattlefield && ev.EntersAsCopyOf != nil
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Label: "test: cancel the Clone's entry once it has chosen",
		})
		if err := g.PutFromHandOntoBattlefieldThenForEffect(clone, game.HandEntryOptions{Controller: me.ID},
			func(_ *game.Game, _ uuid.UUID) error { thenRan++; return nil }); err != nil {
			t.Fatalf("put: %v", err)
		}
	})
	c := copyPrompt(g)
	if c == nil {
		t.Fatal("no copy prompt")
	}
	if err := g.ResolveCopyTarget(c.ID, c.Chooser, bears); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	if thenRan != 1 {
		t.Fatalf("the batch's continuation ran %d times after the cancelled entry, want 1", thenRan)
	}
	if g.Battlefield.Contains(clone) || !me.Hand.Contains(clone) {
		t.Error("a cancelled entry leaves the Clone in hand")
	}
}
