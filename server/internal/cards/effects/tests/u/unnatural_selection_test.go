package testsu

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const unnaturalSelectionOracle = "0be10525-09dc-4970-b495-2db7d4d3c3e7"

func TestUnnaturalSelectionIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(unnaturalSelectionOracle)
	if !ok || spec.Name != "Unnatural Selection" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Unnatural Selection spec = %+v, registered %v", spec, ok)
	}
}

func TestUnnaturalSelectionMakesTheTargetTheChosenTypeUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ench := effectstest.PushCatalogPermanent(g, me.ID, "Unnatural Selection", "Enchantment", unnaturalSelectionOracle, false)
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear Warrior", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, ench, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	got := effectstest.EffectiveSubtypes(t, g, bear)
	if !slices.Contains(got, "Goblin") || slices.Contains(got, "Bear") || slices.Contains(got, "Warrior") {
		t.Errorf("subtypes after the change = %v, want only Goblin", got)
	}
	if !effectstest.HasEffectiveType(t, g, bear, "Creature") {
		t.Error("the creature lost its card type")
	}
	// It ends with the turn.
	effectstest.NextTurnOfMine(t, g)
	if got := effectstest.EffectiveSubtypes(t, g, bear); !slices.Contains(got, "Bear") || slices.Contains(got, "Goblin") {
		t.Errorf("subtypes next turn = %v, want the printed Bear Warrior back", got)
	}
}

func TestUnnaturalSelectionRefusesWallAndNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ench := effectstest.PushCatalogPermanent(g, me.ID, "Unnatural Selection", "Enchantment", unnaturalSelectionOracle, false)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	params := game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(bear)}
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, params); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, ench, 0, params)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PendingOfKind(g, game.PendingChoiceCreatureType)
	if c == nil {
		t.Fatal("no creature-type prompt")
	}
	if err := g.ResolveCreatureTypeChoice(c.ID, me.ID, "Wall"); err == nil {
		t.Error("Wall was accepted")
	}
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	if got := effectstest.EffectiveSubtypes(t, g, bear); !slices.Contains(got, "Elf") {
		t.Errorf("subtypes = %v, want Elf", got)
	}
}
