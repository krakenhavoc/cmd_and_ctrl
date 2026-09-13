package game

import (
	"testing"

	"github.com/google/uuid"
)

// legend_rule_test.go — S27, CR 704.5j.
//
// The rule the issue asked for was "planeswalker uniqueness", which
// Dominaria deleted in 2018 by making planeswalkers legendary. So
// these tests are about the legend rule, and two of them are about
// what it deliberately does NOT do: it does not compare planeswalker
// subtypes, and it does not fire across controllers.

func pushLegendForTest(g *Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{CounterLoyalty: 4},
	})
	return id
}

func legendPromptFor(g *Game, chooser uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceLegendRule && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func onBattlefieldByID(g *Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestLegendRuleAsksTheControllerToChoose(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	a := pushLegendForTest(g, me, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")
	b := pushLegendForTest(g, me, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	ch := legendPromptFor(g, me)
	if ch == nil {
		t.Fatal("no legend-rule prompt was queued")
	}
	if len(ch.PickTargetCards) != 2 {
		t.Fatalf("prompt offered %d permanents, want 2", len(ch.PickTargetCards))
	}
	// Both stay on the battlefield while the prompt is open: a pending
	// choice stops priority, and provisionally removing one would fire
	// leave-the-battlefield triggers for a permanent that never left.
	if !onBattlefieldByID(g, a) || !onBattlefieldByID(g, b) {
		t.Error("a duplicate legend left the battlefield before the choice was answered")
	}
}

func TestResolveLegendRuleKeepsTheChosenOne(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	keep := pushLegendForTest(g, me, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")
	drop := pushLegendForTest(g, me, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")

	g.WithWriteLock(func() { g.runStateChecksLocked() })
	ch := legendPromptFor(g, me)
	if ch == nil {
		t.Fatal("no legend-rule prompt was queued")
	}
	if err := g.ResolveLegendRule(ch.ID, me, keep); err != nil {
		t.Fatalf("ResolveLegendRule: %v", err)
	}

	if !onBattlefieldByID(g, keep) {
		t.Error("the chosen legend was not kept")
	}
	if onBattlefieldByID(g, drop) {
		t.Error("the unchosen legend stayed on the battlefield")
	}
	found := false
	for _, c := range g.Seats[0].Graveyard.Cards {
		if c.InstanceID == drop {
			found = true
		}
	}
	if !found {
		t.Error("the unchosen legend did not reach its owner's graveyard")
	}
	if legendPromptFor(g, me) != nil {
		t.Error("the legend-rule prompt was not drained")
	}
}

// TestLegendRuleDoesNotFireAcrossControllers — the rule is about one
// player's own board (CR 704.5j says "a player controls"), so two
// players may each keep their own copy.
func TestLegendRuleDoesNotFireAcrossControllers(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	pushLegendForTest(g, g.Seats[0].ID, "Sol Ring's Legendary Cousin", "Legendary Artifact")
	pushLegendForTest(g, g.Seats[1].ID, "Sol Ring's Legendary Cousin", "Legendary Artifact")

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	for _, p := range g.Seats {
		if legendPromptFor(g, p.ID) != nil {
			t.Fatalf("a legend-rule prompt fired for %s across two controllers", p.Name)
		}
	}
}

// TestLegendRuleReadsNamesNotPlaneswalkerTypes is the rule the issue
// asked for, corrected. Before Dominaria, CR 704.5k compared
// planeswalker SUBTYPES and these two would not have coexisted; today
// they legally do, and an implementation of the old rule would be
// both more work and wrong.
func TestLegendRuleReadsNamesNotPlaneswalkerTypes(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	a := pushLegendForTest(g, me, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")
	b := pushLegendForTest(g, me, "Teferi, Hero of Dominaria", "Legendary Planeswalker — Teferi")

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if legendPromptFor(g, me) != nil {
		t.Error("two DIFFERENT Teferis triggered the legend rule")
	}
	if !onBattlefieldByID(g, a) || !onBattlefieldByID(g, b) {
		t.Error("two different Teferis cannot coexist")
	}
}

// TestLegendRuleIgnoresNonLegendaryDuplicates — three Forests are
// fine, and so are three tokens of the same nonlegendary name.
func TestLegendRuleIgnoresNonLegendaryDuplicates(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	pushLegendForTest(g, me, "Grizzly Bears", "Creature — Bear")
	pushLegendForTest(g, me, "Grizzly Bears", "Creature — Bear")

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if legendPromptFor(g, me) != nil {
		t.Error("two nonlegendary permanents with the same name triggered the legend rule")
	}
}

// TestLegendRuleQueuesOnlyOnePromptPerGroup — the SBA loop runs the
// check repeatedly until the game is quiet, and a prompt is not an
// answer, so the guard against re-queuing is what stops the loop from
// stacking up 32 copies of the same question.
func TestLegendRuleQueuesOnlyOnePromptPerGroup(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	pushLegendForTest(g, me, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")
	pushLegendForTest(g, me, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")

	g.WithWriteLock(func() {
		g.runStateChecksLocked()
		g.runStateChecksLocked()
	})

	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceLegendRule {
			n++
		}
	}
	if n != 1 {
		t.Errorf("legend-rule prompts queued = %d, want exactly 1", n)
	}
}
