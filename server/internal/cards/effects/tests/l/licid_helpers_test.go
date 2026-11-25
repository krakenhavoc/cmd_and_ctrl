package testsl

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// licid_helpers_test.go — #5781: LicidAbility, LicidForm and an enchant
// creature Aura form, declared the way a licid card will, on a test card.

const licidHelpersOracle = "test-licid-5781"

func TestLicidHelpersBecomeAnAuraAndEndTheEffect(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     licidHelpersOracle,
		Name:         "Test Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       append(effects.LicidForm(0), effects.PumpAttached(1, 1)),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{B}", "{B}")},
	})
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	licid := effectstest.B12Push(g, me.ID, "Test Licid", "Creature — Licid", licidHelpersOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, licid, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	l := effectstest.B12Card(t, g, licid)
	if !l.IsAura() || l.IsCreature() || l.AttachedTo.ID != bear {
		t.Fatalf("the licid is not an Aura on the bear (aura %v, creature %v, attached %+v)", l.IsAura(), l.IsCreature(), l.AttachedTo)
	}
	if p := effectstest.B12Card(t, g, bear).CurrentPower(); p != 3 {
		t.Errorf("the enchanted bear has power %d, want 3", p)
	}
	if err := g.ActivateCatalogAbility(me.ID, licid, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("the licid activated the ability it lost")
	}

	if err := g.PerformSpecialAction(me.ID, licid, game.SpecialActionEndEffect, game.SpecialActionParams{}); err != nil {
		t.Fatalf("end_effect: %v", err)
	}
	l = effectstest.B12Card(t, g, licid)
	if l.IsAura() || !l.IsCreature() || l.IsAttached() {
		t.Errorf("after ending, the licid is not an unattached creature (aura %v, creature %v, attached %+v)", l.IsAura(), l.IsCreature(), l.AttachedTo)
	}
	if p := effectstest.B12Card(t, g, bear).CurrentPower(); p != 2 {
		t.Errorf("after ending, the bear has power %d, want 2", p)
	}
}
