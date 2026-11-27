package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b693_test.go — the eleven Mercadian Masques lands (#5619),
// each played through the real catalog and the engine.

type mmqLand struct{ name, oracle, color string }

var mmqStorage = []mmqLand{
	{"Fountain of Cho", "0799df10-b489-4f79-bf98-7a0c500b46a1", "W"},
	{"Subterranean Hangar", "0bbd5a04-c281-4afb-98a1-657b4eca102c", "B"},
	{"Mercadian Bazaar", "b02ab3c7-fe4a-443c-b860-ba971d3301b0", "R"},
	{"Rushwood Grove", "ccb2f92e-69c0-415c-81cd-52c384b3b233", "G"},
	{"Saprazzan Cove", "f7dda04a-c9c6-4952-9bbc-87e3c7480347", "U"},
}

var mmqDepletion = []mmqLand{
	{"Hickory Woodlot", "26259c65-8f4e-42a6-b8a7-f65c36d35c4d", "G"},
	{"Remote Farm", "2c38f4c7-1b3f-42b4-a175-edab7acd6cc6", "W"},
	{"Peat Bog", "a176924c-78fc-4151-b2b5-1547b1114a40", "B"},
	{"Sandstone Needle", "c8e0a1a5-8188-4677-9d8a-a18eb593343a", "R"},
	{"Saprazzan Skerry", "e4e6e796-39ce-4a63-8c61-c7c956d75d78", "U"},
}

const mmqHenge = "829474df-6413-4323-aef6-f878cb0e797c"

func TestMmqB693Registered(t *testing.T) {
	all := append(append([]mmqLand{}, mmqStorage...), mmqDepletion...)
	all = append(all, mmqLand{"Henge of Ramos", mmqHenge, ""})
	for _, l := range all {
		effectstest.StdRegisteredFull(t, l.oracle, l.name)
	}
}

func TestMmqStorageLandsEnterTappedCountAndPay(t *testing.T) {
	for _, tc := range mmqStorage {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			entered := effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if !effectstest.B12Card(t, g, entered).Tapped {
				t.Fatal("did not enter tapped")
			}
			land := effectstest.B12Push(g, me.ID, tc.name, "Land", tc.oracle, 0, 0)
			// {T}: put a storage counter (a stack ability).
			effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
			if got := effectstest.CounterCount(g, land, game.CounterStorage); got != 1 {
				t.Fatalf("storage counters %d, want 1", got)
			}
			if !effectstest.B16Tapped(t, g, land) {
				t.Error("the counter ability did not tap the land")
			}
			effectstest.AddCountersForTest(g, land, game.CounterStorage, 2)
			effectstest.UntapForTest(g, land)
			effectstest.ActivateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{CounterCounts: []int{3}})
			if got := effectstest.PoolColours(me); got[tc.color] != 3 || me.ManaPool.Total() != 3 {
				t.Errorf("pool %v, want three {%s}", got, tc.color)
			}
			if got := effectstest.CounterCount(g, land, game.CounterStorage); got != 0 {
				t.Errorf("counters left %d, want 0", got)
			}
		})
	}
}

func TestMmqDepletionLandsEnterWithTwoAndSacrificeOnTheLast(t *testing.T) {
	for _, tc := range mmqDepletion {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			land := effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if !effectstest.B12Card(t, g, land).Tapped {
				t.Fatal("did not enter tapped")
			}
			if got := effectstest.CounterCount(g, land, effects.MmqDepletionCounter); got != 2 {
				t.Fatalf("depletion counters %d, want 2", got)
			}
			effectstest.UntapForTest(g, land)
			effectstest.ActivateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
			if got := effectstest.PoolColours(me); got[tc.color] != 2 || me.ManaPool.Total() != 2 {
				t.Fatalf("pool %v, want two {%s}", got, tc.color)
			}
			if !g.Battlefield.Contains(land) {
				t.Fatal("sacrificed with a counter still on it")
			}
			if got := effectstest.CounterCount(g, land, effects.MmqDepletionCounter); got != 1 {
				t.Fatalf("depletion counters %d, want 1", got)
			}
			effectstest.UntapForTest(g, land)
			effectstest.ActivateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
			if got := effectstest.PoolColours(me); got[tc.color] != 4 {
				t.Errorf("pool %v, want four {%s}", got, tc.color)
			}
			if g.Battlefield.Contains(land) {
				t.Error("the land was not sacrificed with no counters left")
			}
		})
	}
}

func TestMmqHengeOfRamosTapsForOneOrPaysTwoForAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.B12Push(g, me.ID, "Henge of Ramos", "Land", mmqHenge, 0, 0)
	effectstest.ActivateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
	if got := effectstest.PoolColours(me); got["C"] != 1 || me.ManaPool.Total() != 1 {
		t.Fatalf("pool %v, want one {C}", got)
	}
	spec, _ := effects.Lookup(mmqHenge)
	if len(spec.ManaAbilities) != 2 || spec.ManaAbilities[1].Cost.Mana != "{2}" || !spec.ManaAbilities[1].Cost.Tap {
		t.Fatalf("second ability is not {2},{T}: %+v", spec.ManaAbilities)
	}
}
