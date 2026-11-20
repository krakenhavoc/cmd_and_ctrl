package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_cost_cards_test.go — #625, the five catalog cards the
// counter-removal cost component made whole (or whole but for an
// unrelated caveat): Heart of Kiran, Dragon's Hoard, Mikaeus, the
// Lunarch, Benevolent Hydra and Fain, the Broker. The component's own
// rules are pinned in game/counter_cost_test.go; these tests are about
// each card doing what its oracle text says.

const (
	HeartOfKiranOracle    = "e2ee410f-2467-4f1f-84a0-8a79faedc0b3"
	CcDragonsHoardOracle  = "8cf77dc4-763b-41b7-a5da-0ef2734f08e6"
	CcMikaeusOracle       = "82f3faa8-39fa-450b-843f-d60a4c36d8f7"
	CcBenevolentHydra     = "01dbf1bc-ca62-4fb6-959c-ef7c0dc03bb0"
	CcFainTheBrokerOracle = "31b990e3-9bad-4b0a-a9d5-f5b9ed2ad0b0"
)

// pushWalkerForCounterCost seats a planeswalker with `loyalty` loyalty
// counters under `owner`'s control.
func PushWalkerForCounterCost(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

// renameBattlefieldForCounterCost gives a battlefield card a distinct
// name, so two pushWalkerForCounterCost walkers under one controller
// are two different legendary permanents rather than a legend-rule
// prompt.
func RenameBattlefieldForCounterCost(g *game.Game, id uuid.UUID, name string) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Name = name
		}
	}
}

func SetCounters(g *game.Game, id uuid.UUID, counters map[string]int) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Counters = counters
		}
	}
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
}
