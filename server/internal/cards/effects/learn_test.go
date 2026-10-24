package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// learn_test.go — the keyword action learn (CR 701.48a, #3525) on the
// outside-the-game seam, driven through Cram Session, whose only other
// clause is a life gain.

const (
	cramSessionOracle = "6e5b8bb2-9dd8-4986-9d0a-659114be5a2a"
	lessonTypeLine    = "Sorcery — Lesson"
)

// lessonCard is a Lesson in a sideboard.
func lessonCard(name string) game.Card { return outsideCard(name, lessonTypeLine) }

// learnTable casts Cram Session for the active seat (seat 0) over a
// sideboard of `side`, and settles the spell. The life gain has
// happened, and the discard prompt is open when the hand has a card.
func learnTable(t *testing.T, side ...game.Card) (*game.Game, *game.Player) {
	t.Helper()
	g := newSideboardCatalogGame(t, 2, side...)
	me := g.Seats[0]
	life := me.Life
	castCatalogSpell(t, g, "Cram Session", "Sorcery", cramSessionOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("Cram Session's life gain: %d -> %d, want +4", life, me.Life)
	}
	return g, me
}

// learnDiscardPrompt is the open "you may discard a card" prompt: a
// choose_cards over the chooser's own hand.
func learnDiscardPrompt(g *game.Game, me *game.Player) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID &&
			len(c.ChooseCards) > 0 && me.Hand.Contains(c.ChooseCards[0]) {
			return c
		}
	}
	return nil
}

// learnLessonPrompt is the open Lesson pick: a choose_cards whose
// candidates are not in the hand (they are outside the game).
func learnLessonPrompt(g *game.Game, me *game.Player) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID &&
			len(c.ChooseCards) > 0 && !me.Hand.Contains(c.ChooseCards[0]) {
			return c
		}
	}
	return nil
}

// "You may discard a card. If you do, draw a card." A player who
// discards has "discarded a card", so the Lesson half is not taken.
func TestLearnDiscardsACardToDrawACard(t *testing.T) {
	g, me := learnTable(t, lessonCard("Environmental Sciences"))
	c := learnDiscardPrompt(g, me)
	if c == nil {
		t.Fatalf("no discard prompt: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want a \"may\" of one", c.ChooseMin, c.ChooseMax)
	}
	handBefore := me.Hand.Size()
	pitched := c.ChooseCards[0]
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pitched}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) {
		t.Error("the chosen card was not discarded")
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d -> %d, want one discarded and one drawn", handBefore, me.Hand.Size())
	}
	if learnLessonPrompt(g, me) != nil {
		t.Error("a player who discarded was also offered a Lesson")
	}
	if len(me.SideboardCards()) != 1 {
		t.Errorf("the Lesson left the sideboard: %v", sideNamed(me))
	}
}

// "If you didn't discard a card, you may reveal a Lesson card you own
// from outside the game and put it into your hand." Only Lessons are
// offered; the choice is revealed; the rest stays outside.
func TestLearnWithNoDiscardRevealsALessonFromOutside(t *testing.T) {
	g, me := learnTable(t,
		lessonCard("Environmental Sciences"), outsideCard("Lightning Bolt", "Instant"), lessonCard("Introduction to Prophecy"))
	opp := g.Seats[1]
	c := learnDiscardPrompt(g, me)
	if c == nil {
		t.Fatalf("no discard prompt: %+v", g.PendingChoices)
	}
	handBefore := me.Hand.Size()
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline the discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	lc := learnLessonPrompt(g, me)
	if lc == nil {
		t.Fatalf("no Lesson prompt after declining the discard: %+v", g.PendingChoices)
	}
	if got := candidateNames(g, lc); !equalNames(got, []string{"Environmental Sciences", "Introduction to Prophecy"}) {
		t.Errorf("candidates %v, want only the Lessons", got)
	}
	if lc.ChooseMin != 0 || lc.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want a \"may\" of one", lc.ChooseMin, lc.ChooseMax)
	}
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{sideCardID(me, "Lightning Bolt")}); err == nil {
		t.Error("learn accepted a card that is not a Lesson")
	}
	if err := g.ResolveChooseCards(lc.ID, opp.ID, []uuid.UUID{sideCardID(me, "Environmental Sciences")}); err == nil {
		t.Error("an opponent answered my learn")
	}
	pick := sideCardID(me, "Introduction to Prophecy")
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("pick the Lesson: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Introduction to Prophecy") || me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d (had %d), has the Lesson %v", me.Hand.Size(), handBefore, outsideHandHas(me, "Introduction to Prophecy"))
	}
	if got := sideNamed(me); !equalNames(got, []string{"Environmental Sciences", "Lightning Bolt"}) {
		t.Errorf("sideboard %v, want the other two to stay outside", got)
	}
	for _, card := range me.Hand.Cards {
		if card.InstanceID == pick && !card.IsKnownTo(opp.ID) {
			t.Error("the revealed Lesson is not known to the opponent")
		}
	}
	revealed := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.CardID == pick && ev.OldZone == game.ZoneOutside {
			revealed = true
		}
	}
	if !revealed {
		t.Error("no reveal event for the Lesson")
	}
}

// Both halves are "may": a player can take neither.
func TestLearnMayTakeNeitherHalf(t *testing.T) {
	g, me := learnTable(t, lessonCard("Environmental Sciences"))
	handBefore := me.Hand.Size()
	if err := g.ResolveChooseCards(learnDiscardPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if err := g.ResolveChooseCards(learnLessonPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore || len(me.SideboardCards()) != 1 || len(g.PendingChoices) != 0 {
		t.Errorf("declining both changed something: hand %d -> %d, sideboard %v, prompts %+v",
			handBefore, me.Hand.Size(), sideNamed(me), g.PendingChoices)
	}
}

// An empty hand has nothing to discard, so the player "didn't discard a
// card" and goes straight to the Lessons without being asked about a
// discard (CR 701.9a: nothing to discard is asked nothing).
func TestLearnWithAnEmptyHandGoesStraightToTheLessons(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, lessonCard("Environmental Sciences"))
	me := g.Seats[0]
	castCatalogSpell(t, g, "Cram Session", "Sorcery", cramSessionOracle, nil)
	me.Hand.Cards = nil
	passPriorityAroundTable(t, g)
	if learnDiscardPrompt(g, me) != nil {
		t.Fatalf("an empty hand was asked to discard: %+v", g.PendingChoices)
	}
	lc := learnLessonPrompt(g, me)
	if lc == nil {
		t.Fatalf("no Lesson prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{sideCardID(me, "Environmental Sciences")}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Environmental Sciences") {
		t.Error("the Lesson did not come into the hand")
	}
}

// With no Lesson outside the game, declining the discard asks nothing
// more: a clause that finds no card is not an error (CR 608.2c).
func TestLearnWithNoLessonOutsideAsksOnlyTheDiscard(t *testing.T) {
	g, me := learnTable(t, outsideCard("Lightning Bolt", "Instant"))
	if err := g.ResolveChooseCards(learnDiscardPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("a prompt went up with no Lesson to find: %+v", g.PendingChoices)
	}
}

// ADR 0198: a deck that seated no sideboard takes its Lesson from the
// card database, minted as a real card it owns.
func TestLearnTakesALessonFromTheDatabaseWithoutASideboard(t *testing.T) {
	prev := game.OutsideCollection
	game.OutsideCollection = testCollection{cards: []game.Card{
		dbCard("Lightning Bolt", "Instant", lightningBoltOracle, "R"),
		dbCard("Environmental Sciences", lessonTypeLine, "", "G"),
		dbCard("Introduction to Annihilation", lessonTypeLine, ""),
	}}
	t.Cleanup(func() { game.OutsideCollection = prev })

	g, me := learnTable(t)
	opp := g.Seats[1]
	if err := g.ResolveChooseCards(learnDiscardPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c := latestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no collection pick: %+v", g.PendingChoices)
	}
	if want := []string{"Environmental Sciences", "Introduction to Annihilation"}; !slices.Equal(c.CollectionSuggestions, want) {
		t.Errorf("suggestions %v, want only the Lessons %v", c.CollectionSuggestions, want)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Lightning Bolt"}); err == nil {
		t.Error("learn took a card that is not a Lesson from the database")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Environmental Sciences"}); err != nil {
		t.Fatalf("ResolveOutsideCollection: %v", err)
	}
	passPriorityAroundTable(t, g)
	card, ok := handCardNamed(me, "Environmental Sciences")
	if !ok {
		t.Fatal("the Lesson did not come into the hand")
	}
	if card.Owner != me.ID || !card.IsKnownTo(opp.ID) {
		t.Errorf("owner %v, revealed to the opponent %v", card.Owner, card.IsKnownTo(opp.ID))
	}
}

// A table with a learn question open is a restore point (#4256): both
// halves are keyed prompts, so the census counts no continuation frame.
// keyed_prompts_restore_test.go answers each one on a restored table.
// Restoring the point before the cast, and casting again, learns as
// normal.
func TestLearnPromptIsARestorePointAndTheGameBeforeItRestores(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, lessonCard("Environmental Sciences"))
	me := g.Seats[0]
	before := g.CaptureSnapshot()
	if !before.Restorable() {
		t.Fatalf("the table before the cast is not a restore point: %+v", before.Continuations)
	}

	castCatalogSpell(t, g, "Cram Session", "Sorcery", cramSessionOracle, nil)
	passPriorityAroundTable(t, g)
	if learnDiscardPrompt(g, me) == nil {
		t.Fatalf("no discard prompt: %+v", g.PendingChoices)
	}
	open := g.CaptureSnapshot()
	if !open.Restorable() || open.Continuations.ChoiceResumeFrames != 0 {
		t.Errorf("a table waiting on a learn discard is not a restore point: %+v", open.Continuations)
	}

	// The last restore point rebuilds the table, sideboard and all, and
	// learn works from it.
	restored := restoreRoundTripFrom(t, before)
	rme := restored.Seats[0]
	if len(rme.SideboardCards()) != 1 || !rme.HasSideboard {
		t.Fatalf("the restored table lost its sideboard: %v", sideNamed(rme))
	}
	castCatalogSpell(t, restored, "Cram Session", "Sorcery", cramSessionOracle, nil)
	passPriorityAroundTable(t, restored)
	c := learnDiscardPrompt(restored, rme)
	if c == nil {
		t.Fatalf("no discard prompt after the restore: %+v", restored.PendingChoices)
	}
	if err := restored.ResolveChooseCards(c.ID, rme.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, restored)
	lc := learnLessonPrompt(restored, rme)
	if lc == nil {
		t.Fatalf("no Lesson prompt after the restore: %+v", restored.PendingChoices)
	}
	if err := restored.ResolveChooseCards(lc.ID, rme.ID, []uuid.UUID{sideCardID(rme, "Environmental Sciences")}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, restored)
	if !outsideHandHas(rme, "Environmental Sciences") {
		t.Error("the Lesson did not come into the hand after the restore")
	}
}

// restoreRoundTripFrom is restoreRoundTrip for a snapshot taken earlier.
func restoreRoundTripFrom(t *testing.T, snap *game.GameSnapshot) *game.Game {
	t.Helper()
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	restored.WithWriteLock(func() { restored.RecomputeLayersIfStaleLocked() })
	return restored
}
