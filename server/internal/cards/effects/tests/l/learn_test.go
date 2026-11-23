package testsl

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// "You may discard a card. If you do, draw a card." A player who
// discards has "discarded a card", so the Lesson half is not taken.
func TestLearnDiscardsACardToDrawACard(t *testing.T) {
	g, me := effectstest.LearnTable(t, effectstest.LessonCard("Environmental Sciences"))
	c := effectstest.LearnDiscardPrompt(g, me)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) {
		t.Error("the chosen card was not discarded")
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d -> %d, want one discarded and one drawn", handBefore, me.Hand.Size())
	}
	if effectstest.LearnLessonPrompt(g, me) != nil {
		t.Error("a player who discarded was also offered a Lesson")
	}
	if len(me.SideboardCards()) != 1 {
		t.Errorf("the Lesson left the sideboard: %v", effectstest.SideNamed(me))
	}
}

// "If you didn't discard a card, you may reveal a Lesson card you own
// from outside the game and put it into your hand." Only Lessons are
// offered; the choice is revealed; the rest stays outside.
func TestLearnWithNoDiscardRevealsALessonFromOutside(t *testing.T) {
	g, me := effectstest.LearnTable(t,
		effectstest.LessonCard("Environmental Sciences"), effectstest.OutsideCard("Lightning Bolt", "Instant"), effectstest.LessonCard("Introduction to Prophecy"))
	opp := g.Seats[1]
	c := effectstest.LearnDiscardPrompt(g, me)
	if c == nil {
		t.Fatalf("no discard prompt: %+v", g.PendingChoices)
	}
	handBefore := me.Hand.Size()
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline the discard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	lc := effectstest.LearnLessonPrompt(g, me)
	if lc == nil {
		t.Fatalf("no Lesson prompt after declining the discard: %+v", g.PendingChoices)
	}
	if got := effectstest.CandidateNames(g, lc); !effectstest.EqualNames(got, []string{"Environmental Sciences", "Introduction to Prophecy"}) {
		t.Errorf("candidates %v, want only the Lessons", got)
	}
	if lc.ChooseMin != 0 || lc.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want a \"may\" of one", lc.ChooseMin, lc.ChooseMax)
	}
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Lightning Bolt")}); err == nil {
		t.Error("learn accepted a card that is not a Lesson")
	}
	if err := g.ResolveChooseCards(lc.ID, opp.ID, []uuid.UUID{effectstest.SideCardID(me, "Environmental Sciences")}); err == nil {
		t.Error("an opponent answered my learn")
	}
	pick := effectstest.SideCardID(me, "Introduction to Prophecy")
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("pick the Lesson: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Introduction to Prophecy") || me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d (had %d), has the Lesson %v", me.Hand.Size(), handBefore, effectstest.OutsideHandHas(me, "Introduction to Prophecy"))
	}
	if got := effectstest.SideNamed(me); !effectstest.EqualNames(got, []string{"Environmental Sciences", "Lightning Bolt"}) {
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
	g, me := effectstest.LearnTable(t, effectstest.LessonCard("Environmental Sciences"))
	handBefore := me.Hand.Size()
	if err := g.ResolveChooseCards(effectstest.LearnDiscardPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolveChooseCards(effectstest.LearnLessonPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore || len(me.SideboardCards()) != 1 || len(g.PendingChoices) != 0 {
		t.Errorf("declining both changed something: hand %d -> %d, sideboard %v, prompts %+v",
			handBefore, me.Hand.Size(), effectstest.SideNamed(me), g.PendingChoices)
	}
}

// An empty hand has nothing to discard, so the player "didn't discard a
// card" and goes straight to the Lessons without being asked about a
// discard (CR 701.9a: nothing to discard is asked nothing).
func TestLearnWithAnEmptyHandGoesStraightToTheLessons(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.LessonCard("Environmental Sciences"))
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	me.Hand.Cards = nil
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LearnDiscardPrompt(g, me) != nil {
		t.Fatalf("an empty hand was asked to discard: %+v", g.PendingChoices)
	}
	lc := effectstest.LearnLessonPrompt(g, me)
	if lc == nil {
		t.Fatalf("no Lesson prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Environmental Sciences")}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Environmental Sciences") {
		t.Error("the Lesson did not come into the hand")
	}
}

// With no Lesson outside the game, declining the discard asks nothing
// more: a clause that finds no card is not an error (CR 608.2c).
func TestLearnWithNoLessonOutsideAsksOnlyTheDiscard(t *testing.T) {
	g, me := effectstest.LearnTable(t, effectstest.OutsideCard("Lightning Bolt", "Instant"))
	if err := g.ResolveChooseCards(effectstest.LearnDiscardPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("a prompt went up with no Lesson to find: %+v", g.PendingChoices)
	}
}

// ADR 0198: a deck that seated no sideboard takes its Lesson from the
// card database, minted as a real card it owns.
func TestLearnTakesALessonFromTheDatabaseWithoutASideboard(t *testing.T) {
	prev := game.OutsideCollection
	game.OutsideCollection = effectstest.TestCollection{Cards: []game.Card{
		effectstest.DbCard("Lightning Bolt", "Instant", effectstest.LightningBoltOracle, "R"),
		effectstest.DbCard("Environmental Sciences", effectstest.LessonTypeLine, "", "G"),
		effectstest.DbCard("Introduction to Annihilation", effectstest.LessonTypeLine, ""),
	}}
	t.Cleanup(func() { game.OutsideCollection = prev })

	g, me := effectstest.LearnTable(t)
	opp := g.Seats[1]
	if err := g.ResolveChooseCards(effectstest.LearnDiscardPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestCollectionPickFor(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.HandCardNamed(me, "Environmental Sciences")
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
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.LessonCard("Environmental Sciences"))
	me := g.Seats[0]
	before := g.CaptureSnapshot()
	if !before.Restorable() {
		t.Fatalf("the table before the cast is not a restore point: %+v", before.Continuations)
	}

	effectstest.CastCatalogSpell(t, g, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LearnDiscardPrompt(g, me) == nil {
		t.Fatalf("no discard prompt: %+v", g.PendingChoices)
	}
	open := g.CaptureSnapshot()
	if !open.Restorable() || open.Continuations.ChoiceResumeFrames != 0 {
		t.Errorf("a table waiting on a learn discard is not a restore point: %+v", open.Continuations)
	}

	// The last restore point rebuilds the table, sideboard and all, and
	// learn works from it.
	restored := effectstest.RestoreRoundTripFrom(t, before)
	rme := restored.Seats[0]
	if len(rme.SideboardCards()) != 1 || !rme.HasSideboard {
		t.Fatalf("the restored table lost its sideboard: %v", effectstest.SideNamed(rme))
	}
	effectstest.CastCatalogSpell(t, restored, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, restored)
	c := effectstest.LearnDiscardPrompt(restored, rme)
	if c == nil {
		t.Fatalf("no discard prompt after the restore: %+v", restored.PendingChoices)
	}
	if err := restored.ResolveChooseCards(c.ID, rme.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, restored)
	lc := effectstest.LearnLessonPrompt(restored, rme)
	if lc == nil {
		t.Fatalf("no Lesson prompt after the restore: %+v", restored.PendingChoices)
	}
	if err := restored.ResolveChooseCards(lc.ID, rme.ID, []uuid.UUID{effectstest.SideCardID(rme, "Environmental Sciences")}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, restored)
	if !effectstest.OutsideHandHas(rme, "Environmental Sciences") {
		t.Error("the Lesson did not come into the hand after the restore")
	}
}
