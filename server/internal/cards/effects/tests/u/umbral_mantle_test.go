package testsu

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const umbralMantleOracle = "44f3241f-af82-4f2e-b15d-7f279bf6e816"

func TestUmbralMantleIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(umbralMantleOracle)
	if !ok || spec.Name != "Umbral Mantle" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Umbral Mantle spec = %+v, registered %v", spec, ok)
	}
}

func TestUmbralMantleGrantsAPumpPaidWithTheUntapSymbol(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	mantle := effectstest.B21Push(g, me.ID, "Umbral Mantle", "Artifact — Equipment", umbralMantleOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, mantle, bear)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	params := game.ActivateAbilityParams{Ref: ref, Strict: true}

	// {Q} needs a tapped creature: an untapped one cannot pay.
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, params); err == nil {
		t.Fatal("activated {Q} on an untapped creature")
	}
	effectstest.TapForTest(t, g, bear)
	effectstest.B16Activate(t, g, me.ID, bear, idx, params)
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature stayed tapped: {Q} did not untap it")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("pumped creature is %d/%d, want 4/4", p, tough)
	}
}

func TestUmbralMantleNeedsThreeManaAndTheAbilityLeavesWithTheMantle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	mantle := effectstest.B21Push(g, me.ID, "Umbral Mantle", "Artifact — Equipment", umbralMantleOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, mantle, bear)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	effectstest.TapForTest(t, g, bear)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err == nil {
		t.Fatal("activated with two mana")
	}
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("a refused activation untapped the creature")
	}
	// Moved off the creature, the ability goes with the Mantle.
	other := effectstest.B16Creature(g, me.ID, "Other", "Creature — Bear", 1, 1, "G")
	effectstest.EquipTo(t, g, me.ID, mantle, other)
	if i, _ := effectstest.GrantedActivatedIndex(t, g, bear); i >= 0 {
		t.Error("the first creature kept the granted ability after the Mantle moved")
	}
	if i, _ := effectstest.GrantedActivatedIndex(t, g, other); i < 0 {
		t.Error("the new creature did not get the granted ability")
	}
}
