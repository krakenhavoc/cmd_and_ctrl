package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_color_of_choice_test.go — #6210: "produces mana of a color of your
// choice instead". CR 106.12b: the replacement "modifies the mana
// production event while such an ability is resolving and producing
// mana", and CR 605.3b: the ability resolves at once, so the colour is
// chosen as part of the tap. Pulse of Llanowar and Harvest Mage are their
// batches'; the shapes are registered here.

const pulseTestOracle = "test-6210-pulse"

func pulseTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: pulseTestOracle,
		Name:     "Pulse Test",
		Replacements: []game.ReplacementEffect{effects.TappedForManaProducesAColorOfYourChoice(
			"If a basic land you control is tapped for mana, it produces mana of a color of your choice instead of any other type.",
			false, effects.BasicLandYouControl)},
	})
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Pulse Test", "Enchantment", pulseTestOracle, false)
	effectstest.ToMainForCost(t, g)
	return g, me, opp
}

func poolColors(p *game.Player) map[string]int {
	out := map[string]int{}
	for _, tok := range p.ManaPool {
		out[tok.Color]++
	}
	return out
}

// The issue's acceptance: a Swamp tapped by the auto-tapper while paying
// {G} produces {G}.
func TestPulseOfLlanowarSwampPaysAGreenCost(t *testing.T) {
	g, me, _ := pulseTable(t)
	swamp := effectstest.FreeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	elf := effectstest.HandCardFull(me, "Green Test", "Sorcery", "{G}", "", []string{"G"})
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("casting a {G} spell with only a Swamp under Pulse: %v", err)
	}
	if !g.Stack.Contains(elf) {
		t.Fatal("the spell should be on the stack")
	}
	if c, _ := g.LookupCardForEffect(swamp); !c.Tapped {
		t.Error("the Swamp should have been tapped to pay")
	}
}

// By hand, the tap asks the five-colour pick a Birds of Paradise asks,
// and the colour picked is the mana.
func TestPulseOfLlanowarHandTapAsksForAColour(t *testing.T) {
	g, me, _ := pulseTable(t)
	swamp := effectstest.FreeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateManaAbility(me.ID, swamp, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping the Swamp: %v", err)
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID)
	if pick == nil {
		t.Fatalf("no colour pick after tapping the Swamp; pool %v", poolColors(me))
	}
	if len(pick.ColorOptions) != 5 {
		t.Fatalf("pick options %v, want the five colours", pick.ColorOptions)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("picking blue: %v", err)
	}
	if got := poolColors(me); got["U"] != 1 || len(got) != 1 {
		t.Errorf("pool %v, want one {U}", got)
	}
}

// Named up front, the colour needs no pick.
func TestPulseOfLlanowarUpfrontColour(t *testing.T) {
	g, me, _ := pulseTable(t)
	swamp := effectstest.FreeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateManaAbility(me.ID, swamp, 0, game.ManaAbilityParams{Colors: []string{"W"}}); err != nil {
		t.Fatalf("tapping the Swamp for white: %v", err)
	}
	if got := poolColors(me); got["W"] != 1 || len(got) != 1 {
		t.Errorf("pool %v, want one {W}", got)
	}
}

// "A basic land YOU control": another player's Swamp still makes {B}.
func TestPulseOfLlanowarLeavesOtherPlayersLandsAlone(t *testing.T) {
	g, _, opp := pulseTable(t)
	swamp := effectstest.FreeSpellPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateManaAbility(opp.ID, swamp, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping the Swamp: %v", err)
	}
	if got := poolColors(opp); got["B"] != 1 || len(got) != 1 {
		t.Errorf("pool %v, want one {B}", got)
	}
}

// Harvest Mage's until-end-of-turn form: one mana of the chosen colour
// instead of any other type and amount, for the player who activated it.
func TestHarvestMageMakesOneManaOfTheChosenColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	src := uuid.New()
	g.WithWriteLock(func() {
		if !g.LandsYouTapProduceOneManaOfAColorOfYourChoiceThisTurnForEffect(src, me.ID, "Harvest Test") {
			t.Fatal("the replacement was not registered")
		}
	})
	swamp := effectstest.FreeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateManaAbility(me.ID, swamp, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tapping the Swamp for red: %v", err)
	}
	if got := poolColors(me); got["R"] != 1 || len(got) != 1 {
		t.Errorf("pool %v, want one {R}", got)
	}
}
