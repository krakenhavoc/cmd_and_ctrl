package testsru

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rules_batch_3946_test.go — #3946: Alela and Trygon Predator name "that
// player" through ControlledByThatPlayer, so with two players hit in one
// turn each trigger offers only ITS victim's permanents. The old
// predicate read the per-turn tally of every player hit and offered both.

// Trygon Predator hit the third player earlier this turn; the trigger
// for the second player must not offer the third player's artifact.
func TestTrygonPredatorOffersOnlyThisTriggersVictimAfterTwoHits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	predator := effectstest.B12Push(g, me.ID, "Trygon Predator", "Creature — Beast", effectstest.B32TrygonPredatorOracle, 2, 3)
	rock := effectstest.B12Permanent(g, victim.ID, "Their Rock", "Artifact")
	otherRock := effectstest.B12Permanent(g, other.ID, "Other Rock", "Artifact")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: me.ID, Source: predator,
			Target: other.ID, Amount: 2, Combat: true,
		})
	})

	// The first hit's own trigger: answered and resolved, so the second
	// hit is the one under test.
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, otherRock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AttackWith(t, g, victim.ID, predator)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, rock) {
		t.Error("the player this trigger's damage hit has their artifact offered")
	}
	if effectstest.HasID(p.PickTargetCards, otherRock) {
		t.Error("a player hit earlier this turn, not by this trigger's damage, is not offered")
	}
}

// Alela's Faeries hit the third player earlier this turn; the goad for
// the second player must not offer the third player's creature.
func TestAlelaOffersOnlyThisTriggersVictimAfterTwoHits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	alela := effectstest.B12Push(g, me.ID, "Alela, Cunning Conqueror", "Legendary Creature — Faerie Warlock", effectstest.B33AlelaOracle, 2, 4)
	theirs := effectstest.B12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	otherBear := effectstest.B12Creature(g, other.ID, "Other Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: me.ID, Source: alela,
			Target: other.ID, Amount: 2, Combat: true,
		})
	})

	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, otherBear)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AttackWith(t, g, victim.ID, alela)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("the player this trigger's damage hit has their creature offered")
	}
	if effectstest.HasID(p.PickTargetCards, otherBear) {
		t.Error("a player hit earlier this turn, not by this trigger's damage, is not offered")
	}
}
