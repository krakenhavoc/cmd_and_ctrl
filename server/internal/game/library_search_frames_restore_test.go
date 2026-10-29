package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// library_search_frames_restore_test.go — #4361 S4: the library-order and
// search frames as restore points. A search prompt holds only data, an
// outside pick is always keyed (a pick with no key of its own rides
// outsidePickPlain), and a discard batch nothing follows orders the cards
// a replacement put on top with a zero tail. Each test captures the table
// mid-prompt, restores it strictly from JSON, and answers on the restored
// table.

func strictRestoreOf(t *testing.T, g *Game) *Game {
	t.Helper()
	snap := throughJSON(t, g.CaptureSnapshot())
	if !snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 0 {
		t.Fatalf("the open prompt is not a restore point: %#v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	return restored
}

// Wish's "name a card you own from outside the game to play" names no
// pick of its own: it mints the card into the container, where the
// standing right covers it. It rides outsidePickPlain, so the open name
// prompt is a restore point, and the restored prompt mints the card.
func TestAPlainCollectionPickIsARestorePoint(t *testing.T) {
	standardCollection(t)
	g := newSideboardGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if !g.QueueOutsidePickForEffect(OutsidePickPrompt{Player: me.ID, Question: "Wish", Destination: ChoosePlayFromOutside}) {
			t.Fatal("no collection pick")
		}
	})
	c := g.PendingChoices[0]
	if c.Kind != PendingChoiceOutsideCollection || c.PickThen != outsidePickPlain.Key() {
		t.Fatalf("prompt %s with pick %q, want a collection pick on %q", c.Kind, c.PickThen, outsidePickPlain.Key())
	}

	g = strictRestoreOf(t, g)
	me = g.Seats[0]
	c = g.PendingChoices[0]
	if found, err := g.SearchOutsideCollection(c.ID, me.ID, "bolt", 0); err != nil || len(found) != 1 {
		t.Fatalf("the restored prompt's search found %d (%v), want Lightning Bolt", len(found), err)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Lightning Bolt"}); err != nil {
		t.Fatalf("ResolveOutsideCollection on the restored table: %v", err)
	}
	if side := me.SideboardCards(); len(side) != 1 || side[0].Name != "Lightning Bolt" {
		t.Fatalf("the container holds %v, want the minted Lightning Bolt", side)
	}
}

// A sideboard pick with no key of its own rides the same plain pick: a
// keyed choose_cards prompt, a restore point.
func TestAPlainSideboardPickIsARestorePoint(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"), sbCard("Lightning Bolt", "Instant"))
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if !g.QueueOutsidePickForEffect(OutsidePickPrompt{Player: me.ID, Question: "Wish", Max: 1}) {
			t.Fatal("no sideboard pick")
		}
	})
	c := g.PendingChoices[0]
	if c.Kind != PendingChoiceChooseCards || c.PickThen != outsidePickPlain.Key() {
		t.Fatalf("prompt %s with pick %q, want a keyed choose_cards", c.Kind, c.PickThen)
	}
	g = strictRestoreOf(t, g)
	me = g.Seats[0]
	c = g.PendingChoices[0]
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{c.ChooseCards[0]}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if len(g.PendingChoices) != 0 || len(me.SideboardCards()) != 2 {
		t.Fatalf("after the answer: %d prompts, %d cards outside; want none and both", len(g.PendingChoices), len(me.SideboardCards()))
	}
}

// A test outside pick with a filter, a set rule and a continuation asks
// and answers through its key on a restored table.
func TestAKeyedCollectionPickWithASetRuleIsARestorePoint(t *testing.T) {
	standardCollection(t)
	g := newSideboardGame(t)
	me := g.Seats[0]
	var got []uuid.UUID
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(outsideHookedTwoNames(OutsidePickPrompt{Player: me.ID, Min: 0, Max: 2},
			func(_ *Game, picked []uuid.UUID) error { got = picked; return nil }))
	})
	g = strictRestoreOf(t, g)
	me = g.Seats[0]
	c := g.PendingChoices[0]
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Lightning Bolt", "Lightning Bolt"}); err == nil {
		t.Fatal("the restored set rule accepted two cards with one name")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Lightning Bolt", "Llanowar Elves"}); err != nil {
		t.Fatalf("ResolveOutsideCollection: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("the continuation was handed %v, want the two minted cards", got)
	}
}

// CR 401.4 (#2949): a cost discard whose commanders' owners answered
// before the payment walks the closure batch, and nothing follows it. The
// cards a replacement put on top are ordered with a zero tail, so the
// ordering prompt is a restore point.
func TestACostDiscardsTopOrderingIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	g.WithWriteLock(func() { p.Hand.Cards = nil })
	first := handCardForDiscard(t, g, p, "First")
	second := handCardForDiscard(t, g, p, "Second")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(discardReplacement("to the library instead", ZoneLibrary, false, DiscardCauseCost))
		// A non-nil answer map is what a cost discard carries once a
		// pitched commander's owner has answered (#1397): the closure batch.
		if err := g.discardCardsLocked(p.ID, []uuid.UUID{first, second}, discardOptions{
			cause: DiscardCauseCost, commanderAnswers: map[uuid.UUID]bool{},
		}); err != nil {
			t.Fatalf("discardCardsLocked: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoicePutInLibrary {
		t.Fatalf("pending %v, want the ordering prompt", g.PendingChoices)
	}

	g = strictRestoreOf(t, g)
	p = g.Seats[1]
	c := g.PendingChoices[0]
	want := []uuid.UUID{first, second}
	if err := g.ResolvePutInLibrary(c.ID, p.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary on the restored table: %v", err)
	}
	var top []uuid.UUID
	for i := len(p.Library.Cards) - 1; i >= 0 && len(top) < 2; i-- {
		top = append(top, p.Library.Cards[i].InstanceID)
	}
	if !slices.Equal(top, want) {
		t.Fatalf("library top %v, want the chosen %v", top, want)
	}
}
