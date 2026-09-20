package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const yawgmothOracle = "a1e232c0-dc38-47be-a5a0-f68bc1d86a29"

func TestYawgmothIsFullyImplemented(t *testing.T) {
	spec, ok := Lookup(yawgmothOracle)
	if !ok {
		t.Fatal("Yawgmoth, Thran Physician is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness = %v, caveats = %q, want CompletenessFull with none", spec.Completeness, spec.Caveats)
	}
}

func yawgmothPush(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Yawgmoth, Thran Physician",
		TypeLine: "Legendary Creature — Human Cleric", OracleID: yawgmothOracle,
		Power: 2, Toughness: 4, Owner: owner, Controller: owner,
	})
}

// TestYawgmothPaysLifeAndSacrificesForACounterAndADraw exercises the
// first ability end to end: 1 life paid, another creature sacrificed
// (never itself), a -1/-1 counter on the chosen target, and a card
// drawn.
func TestYawgmothPaysLifeAndSacrificesForACounterAndADraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	yawgmoth := yawgmothPush(g, me.ID)
	fodder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	victim := seedBear(g, me.ID)

	if err := g.ActivateCatalogAbility(me.ID, yawgmoth, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{yawgmoth}}); err == nil {
		t.Fatal("Yawgmoth can't sacrifice himself")
	}

	beforeLife, beforeHand := me.Life, me.Hand.Size()
	b16Activate(t, g, me.ID, yawgmoth, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      cardRefs(victim),
	})
	if got := me.Life; got != beforeLife-1 {
		t.Errorf("life %d, want %d (pay 1 life)", got, beforeLife-1)
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the sacrificed creature is in the graveyard")
	}
	if got := counterCount(g, victim, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters on the target = %d, want 1", got)
	}
	if got := me.Hand.Size(); got != beforeHand+1 {
		t.Errorf("hand %d, want %d (draw a card)", got, beforeHand+1)
	}
}

// TestYawgmothDrawsEvenWithNoTargetChosen — "up to one" and the draw
// is unconditional: with no target named, the counter never lands but
// the draw still happens.
func TestYawgmothDrawsEvenWithNoTargetChosen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	yawgmoth := yawgmothPush(g, me.ID)
	fodder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})

	beforeHand := me.Hand.Size()
	b16Activate(t, g, me.ID, yawgmoth, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if got := me.Hand.Size(); got != beforeHand+1 {
		t.Errorf("the draw happens whether or not a target was chosen: hand %d, want %d", got, beforeHand+1)
	}
}

// TestYawgmothDiscardsAndProliferates is the second ability: {B}{B},
// discard a card: proliferate.
func TestYawgmothDiscardsAndProliferates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	yawgmoth := yawgmothPush(g, me.ID)
	bear := seedBear(g, me.ID)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 1) })
	discarded := me.Hand.Cards[0].InstanceID

	b06AddMana(me, "B", "B")
	b16Activate(t, g, me.ID, yawgmoth, 1, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discarded}})
	if !me.Graveyard.Contains(discarded) {
		t.Error("the discarded card is in the graveyard")
	}
	if got := counterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("+1/+1 counters after proliferate = %d, want 2", got)
	}
}
