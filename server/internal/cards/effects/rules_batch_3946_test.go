package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rules_batch_3946_test.go — #3946: Alela and Trygon Predator name "that
// player" through ControlledByThatPlayer, so with two players hit in one
// turn each trigger offers only ITS victim's permanents. The old
// predicate read the per-turn tally of every player hit and offered both.

// Trygon Predator hit the third player earlier this turn; the trigger
// for the second player must not offer the third player's artifact.
func TestTrygonPredatorOffersOnlyThisTriggersVictimAfterTwoHits(t *testing.T) {
	g := newCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	predator := b12Push(g, me.ID, "Trygon Predator", "Creature — Beast", b32TrygonPredatorOracle, 2, 3)
	rock := b12Permanent(g, victim.ID, "Their Rock", "Artifact")
	otherRock := b12Permanent(g, other.ID, "Other Rock", "Artifact")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: me.ID, Source: predator,
			Target: other.ID, Amount: 2, Combat: true,
		})
	})

	// The first hit's own trigger: answered and resolved, so the second
	// hit is the one under test.
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, otherRock)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	attackWith(t, g, victim.ID, predator)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, rock) {
		t.Error("the player this trigger's damage hit has their artifact offered")
	}
	if hasID(p.PickTargetCards, otherRock) {
		t.Error("a player hit earlier this turn, not by this trigger's damage, is not offered")
	}
}

// Alela's Faeries hit the third player earlier this turn; the goad for
// the second player must not offer the third player's creature.
func TestAlelaOffersOnlyThisTriggersVictimAfterTwoHits(t *testing.T) {
	g := newCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	alela := b12Push(g, me.ID, "Alela, Cunning Conqueror", "Legendary Creature — Faerie Warlock", b33AlelaOracle, 2, 4)
	theirs := b12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	otherBear := b12Creature(g, other.ID, "Other Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: me.ID, Source: alela,
			Target: other.ID, Amount: 2, Combat: true,
		})
	})

	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, otherBear)
	passPriorityAroundTable(t, g)

	attackWith(t, g, victim.ID, alela)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, theirs) {
		t.Error("the player this trigger's damage hit has their creature offered")
	}
	if hasID(p.PickTargetCards, otherBear) {
		t.Error("a player hit earlier this turn, not by this trigger's damage, is not offered")
	}
}
