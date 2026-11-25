package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// defined_x_cost_test.go — #6240, CR 107.3c: "If a spell or activated
// ability has an {X} … in its cost and/or its text, and the value of X
// is defined by the text of that spell or ability, then that's the value
// of X … The controller of that spell or ability doesn't get to choose
// the value." Soul Foundry's "{X}, {T}: … X is the mana value of that
// card", on a test oracle whose ability gains 1 life.

const testDefinedXOracle = "test-6240-foundry"

func registerDefinedX(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     testDefinedXOracle,
		Name:         "Test Foundry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:  "{X}, {T}: You gain 1 life. X is the mana value of the exiled card.",
			Cost:   effects.Plus(effects.XIsTheManaValueOfTheExiledCard(), effects.TapCost()),
			Effect: effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
}

// foundryWithImprint puts the Foundry on the battlefield and, when mv >=
// 0, a card of that mana value in exile linked to it (CR 607.2a).
func foundryWithImprint(t *testing.T, g *game.Game, me *game.Player, cost string) uuid.UUID {
	t.Helper()
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Foundry", TypeLine: "Artifact",
		OracleID: testDefinedXOracle, Owner: me.ID, Controller: me.ID})
	if cost == "" {
		return id
	}
	src, _ := effectstest.BattlefieldCardByID(g, id)
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Imprinted", TypeLine: "Creature — Golem", ManaCost: cost,
			Owner: me.ID, Controller: me.ID, ExiledWith: game.PermissionCardRef{ID: id, Epoch: src.ObjectEpoch}})
	})
	return id
}

// A 4-mana-value card exiled with it: the ability costs {4} and {T}, the
// row says so and asks for no X, three mana can't pay, and an announced X
// is refused.
func TestDefinedXCostsTheManaValueOfTheExiledCard(t *testing.T) {
	registerDefinedX(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := foundryWithImprint(t, g, me, "{2}{U}{U}")

	price, err := g.PriceActivation(me.ID, id, 0, nil)
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	if price.Ability.Cost.DemandsX() {
		t.Fatal("the ability asks for X, which the text defines (CR 107.3c)")
	}
	if got := price.Total.String(); got != "{4}" {
		t.Fatalf("priced at %s, want {4}", got)
	}

	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("three mana paid a {4} cost")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{XValue: 2, Strict: true}); err == nil {
		t.Fatal("an announced X was accepted for a cost whose X the text defines")
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for {4}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("life = %d, want %d", me.Life, life+1)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Fatalf("%d mana left in the pool, want all four spent", n)
	}
}

// Nothing exiled with it: X is 0, and the ability costs only {T}.
func TestDefinedXIsZeroWithNothingExiled(t *testing.T) {
	registerDefinedX(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := foundryWithImprint(t, g, me, "")
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with nothing exiled: %v", err)
	}
}
