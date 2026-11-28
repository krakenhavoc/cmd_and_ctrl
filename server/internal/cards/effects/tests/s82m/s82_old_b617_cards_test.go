package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b617_cards_test.go — S82 OLD batch 617 (#5607): Eighth Edition
// Auras. Every test plays the real catalog card through the engine.

const (
	oldB617Fear      = "355bbe9b-59bf-470f-8600-410af4c7fe18"
	oldB617Regen     = "89390a33-b289-4edd-a114-6616e49a49c2"
	oldB617Algae     = "a08189f2-1643-496a-8a91-74864528b4db"
	oldB617Steal     = "cd8ae9f2-edac-473a-8846-c08219e617c3"
	oldB617Dehydrate = "db27c686-e202-4b60-9a10-0a0fef25576c"
	oldB617Reflexes  = "dc87b0a5-3d9d-44eb-b415-b022acd63cf1"
	oldB617Terrain   = "ffa40f10-4efb-48d8-b16e-8a5711fcfad1"
)

func TestOldB617CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB617Fear: "Fear", oldB617Regen: "Regeneration", oldB617Algae: "Spreading Algae",
		oldB617Steal: "Steal Artifact", oldB617Dehydrate: "Dehydration", oldB617Reflexes: "Reflexes",
		oldB617Terrain: "Living Terrain",
	} {
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

func TestOldB617FearAndReflexesGrantTheirKeyword(t *testing.T) {
	for _, c := range []struct{ name, oracle, kw string }{
		{"Fear", oldB617Fear, "fear"}, {"Reflexes", oldB617Reflexes, "first strike"},
	} {
		g := effectstest.NewCatalogGame(t)
		bear := effectstest.B12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
		if effectstest.HasEffectiveKeyword(t, g, bear, c.kw) {
			t.Fatalf("%s: the bear starts with %s", c.name, c.kw)
		}
		aura := effectstest.Enchant(t, g, c.name, c.oracle, bear)
		if !effectstest.HasEffectiveKeyword(t, g, bear, c.kw) {
			t.Errorf("%s: the enchanted creature has no %s", c.name, c.kw)
		}
		effectstest.P1Destroy(t, g, aura)
		if effectstest.HasEffectiveKeyword(t, g, bear, c.kw) {
			t.Errorf("%s: the keyword stayed after the Aura left", c.name)
		}
	}
}

func TestOldB617AurasRefuseAnIllegalHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.PushLandFor(g, g.Seats[0].ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, g, "Fear", effectstest.AuraTypeLine, oldB617Fear, effectstest.B16TargetCard(land)); err == nil {
		t.Error("Fear was cast on a land")
	}
}

func TestOldB617RegenerationRegeneratesTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Regeneration", oldB617Regen, bear)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	effectstest.DcMana(t, g, me.ID, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the enchanted creature")
	}
}

func TestOldB617RegenerationNeedsGreenMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Regeneration", oldB617Regen, bear)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with an empty pool")
	}
	effectstest.P1Destroy(t, g, bear)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived with no shield")
	}
}

func TestOldB617StealArtifactTakesControlUntilTheAuraLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	aura := effectstest.Enchant(t, g, "Steal Artifact", oldB617Steal, rock)
	if got := effectstest.ControllerOf(t, g, rock); got != me.ID {
		t.Fatalf("controller %s, want the caster", got)
	}
	effectstest.P1Destroy(t, g, aura)
	if got := effectstest.ControllerOf(t, g, rock); got != opp.ID {
		t.Error("control did not revert when the Aura left")
	}
}

func TestOldB617StealArtifactRefusesANonArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Steal Artifact", effectstest.AuraTypeLine, oldB617Steal, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("Steal Artifact was cast on a non-artifact")
	}
}

func TestOldB617DehydrationKeepsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Dehydration", oldB617Dehydrate, bear)
	effectstest.TapCard(g, bear)
	effectstest.B39NextTurnOf(t, g, 1)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestOldB617DehydrationLeavesOtherCreaturesUntapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.Enchant(t, g, "Dehydration", oldB617Dehydrate, bear)
	effectstest.TapCard(g, bear)
	effectstest.TapCard(g, other)
	effectstest.B39NextTurnOf(t, g, 1)
	if effectstest.B16Tapped(t, g, other) {
		t.Error("an unenchanted creature stayed tapped")
	}
}

func TestOldB617LivingTerrainMakesThe56GreenTreefolkLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.PushLandFor(g, g.Seats[0].ID, "Forest", "Basic Land — Forest")
	aura := effectstest.Enchant(t, g, "Living Terrain", oldB617Terrain, land)
	types := effectstest.EffectiveTypes(t, g, land)
	if !effectstest.ContainsString(types, "Land") || !effectstest.ContainsString(types, "Creature") {
		t.Errorf("types %v, want Land and Creature", types)
	}
	if subs := effectstest.EffectiveSubtypes(t, g, land); !effectstest.ContainsString(subs, "Treefolk") {
		t.Errorf("subtypes %v, want Treefolk", subs)
	}
	if p, tough := effectstest.EffectivePower(t, g, land), effectstest.EffectiveToughness(t, g, land); p != 5 || tough != 6 {
		t.Errorf("P/T %d/%d, want 5/6", p, tough)
	}
	if cols := effectstest.EffectiveColors(t, g, land); len(cols) != 1 || cols[0] != "G" {
		t.Errorf("colors %v, want [G]", cols)
	}
	effectstest.P1Destroy(t, g, aura)
	if types := effectstest.EffectiveTypes(t, g, land); effectstest.ContainsString(types, "Creature") {
		t.Errorf("types %v after the Aura left, want a plain land", types)
	}
}

func oldB617Swamp(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushLandFor(g, owner, "Swamp", "Basic Land — Swamp")
}

func TestOldB617SpreadingAlgaeEnchantsOnlyASwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	forest := effectstest.PushLandFor(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, g, "Spreading Algae", effectstest.AuraTypeLine, oldB617Algae, effectstest.B16TargetCard(forest)); err == nil {
		t.Error("Spreading Algae was cast on a Forest")
	}
}

func TestOldB617SpreadingAlgaeDestroysTheSwampWhenTappedAndReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swamp := oldB617Swamp(g, opp.ID)
	aura := effectstest.Enchant(t, g, "Spreading Algae", oldB617Algae, swamp)
	if !g.Battlefield.Contains(swamp) {
		t.Fatal("the Swamp left before it was tapped")
	}
	if err := g.ActivateManaAbility(opp.ID, swamp, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(swamp) {
		t.Error("the enchanted Swamp survived being tapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Errorf("Spreading Algae is in %v, want its owner's hand", effectstest.B12ZoneOf(g, aura))
	}
}

func TestOldB617SpreadingAlgaeReturnsToHandWhenItIsDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	swamp := oldB617Swamp(g, g.Seats[1].ID)
	aura := effectstest.Enchant(t, g, "Spreading Algae", oldB617Algae, swamp)
	effectstest.P1Destroy(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Errorf("Spreading Algae is in %v, want its owner's hand", effectstest.B12ZoneOf(g, aura))
	}
	if !g.Battlefield.Contains(swamp) {
		t.Error("the Swamp was destroyed with no tap")
	}
}
