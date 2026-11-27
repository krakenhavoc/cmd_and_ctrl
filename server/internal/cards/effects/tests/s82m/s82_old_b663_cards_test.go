package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b663_cards_test.go — Invasion lands, S82 OLD batch 663 (#5616).
// Every test plays the real catalog entry through the engine.

const (
	b663Dig        = "cf438848-da86-4db6-b3b8-4dd8570be3b8"
	b663Necropolis = "ea4d6fcd-21e0-4e9f-b406-a89042998d98"
)

var b663SacrificeLands = []struct{ name, oracle, tap, cash string }{
	{"Ancient Spring", "2160849a-184f-4a49-8931-fd021e16f7cb", "U", "WB"},
	{"Sulfur Vent", "bc2906cd-9e0c-4aa8-b656-fb1045d810ac", "B", "UR"},
	{"Irrigation Ditch", "e0c3c87b-83e3-4cff-b05f-78a192e2bba4", "W", "GU"},
	{"Tinder Farm", "af0ad159-8264-4526-9a95-eddd32c0a13f", "G", "RW"},
	{"Geothermal Crevice", "83f9d84a-a011-4fa9-a3a4-51eb39159c54", "R", "BG"},
}

func TestB663EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{b663Dig: "Archaeological Dig", b663Necropolis: "Keldon Necropolis"}
	for _, l := range b663SacrificeLands {
		all[l.oracle] = l.name
	}
	for oracle, name := range all {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestB663SacrificeLandsEnterTappedTapForOneAndCashInForTwoOthers(t *testing.T) {
	for _, tc := range b663SacrificeLands {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			land := effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if !effectstest.B12Card(t, g, land).Tapped {
				t.Fatal("did not enter tapped")
			}
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.UntapForTest(g, land)
			effectstest.ActivateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
			if got := effectstest.PoolColours(me); got[tc.tap] != 1 || me.ManaPool.Total() != 1 {
				t.Fatalf("tap: pool %v, want one {%s}", got, tc.tap)
			}
			if !g.Battlefield.Contains(land) {
				t.Fatal("the plain tap sacrificed the land")
			}
			effectstest.UntapForTest(g, land)
			effectstest.ActivateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{})
			got := effectstest.PoolColours(me)
			if got[tc.tap] != 1 || got[tc.cash[:1]] != 1 || got[tc.cash[1:]] != 1 || me.ManaPool.Total() != 3 {
				t.Errorf("sacrifice: pool %v, want the earlier {%s} plus {%s}{%s}", got, tc.tap, tc.cash[:1], tc.cash[1:])
			}
			if g.Battlefield.Contains(land) {
				t.Error("the land was not sacrificed")
			}
		})
	}
}

func TestB663ArchaeologicalDigTapsForColorlessAndCashesInForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.B12PlayFromHand(t, g, "Archaeological Dig", "Land", b663Dig, game.CastSpellParams{})
	if effectstest.B12Card(t, g, land).Tapped {
		t.Fatal("entered tapped")
	}
	effectstest.ActivateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
	if got := effectstest.PoolColours(me); got["C"] != 1 || me.ManaPool.Total() != 1 {
		t.Fatalf("tap: pool %v, want one {C}", got)
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("the plain tap sacrificed the land")
	}
	effectstest.UntapForTest(g, land)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("cash in: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("no color prompt")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.PoolColours(me); got["R"] != 1 || got["C"] != 1 || me.ManaPool.Total() != 2 {
		t.Errorf("pool %v, want {C} and {R}", got)
	}
}

func TestB663KeldonNecropolisPingsForASacrificeAndTapsForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Keldon Necropolis", "Legendary Land", b663Necropolis, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	effectstest.AdvanceToMain(t, g)
	target := effectstest.B16TargetPlayer(opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.ApaMana(me, "C", "C", "C", "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("activated without sacrificing a creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("sacrificed the land, which is not a creature")
	}
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{Targets: target, SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if !g.Battlefield.Contains(land) || !effectstest.B16Tapped(t, g, land) {
		t.Error("the land should be tapped and still on the battlefield")
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want -2", life, opp.Life)
	}
	effectstest.UntapForTest(g, land)
	effectstest.ActivateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
	if got := effectstest.PoolColours(me); got["C"] != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("tap: pool %v, want one {C}", got)
	}
}
