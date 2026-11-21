package testsr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const rmsTitanicOracle = "6754753d-790e-4a95-a150-76741c4a02d4"

// TestRMSTitanicSacrificesAndMakesThatManyTreasures — "that many" is
// the damage actually dealt, and the sacrifice and the tokens both
// happen off the one trigger.
func TestRMSTitanicSacrificesAndMakesThatManyTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	titanic := effectstest.PushVehicleForTest(g, me.ID, "RMS Titanic", rmsTitanicOracle, 6, 6)
	effectstest.CrewForTest(t, g, me.ID, titanic, effectstest.PushCrewerForTest(g, me.ID, "Crewer", 3))
	effectstest.AssertKeywords(t, g, titanic, "flying", "trample")

	effectstest.DealCombatDamageToPlayer(g, titanic, opp.ID, 6)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(titanic) {
		t.Error("RMS Titanic should have sacrificed itself")
	}
	if n := effectstest.B43TokensNamed(g, me.ID, "Treasure"); n != 6 {
		t.Fatalf("Treasures = %d, want 6 (the damage dealt)", n)
	}
}

// TestRMSTitanicNoncombatDamageIsNotTheTrigger — the ability is
// explicitly about COMBAT damage; a noncombat hit for the same amount
// must not sacrifice it or make any Treasures.
func TestRMSTitanicNoncombatDamageIsNotTheTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	titanic := effectstest.PushVehicleForTest(g, me.ID, "RMS Titanic", rmsTitanicOracle, 6, 6)
	effectstest.CrewForTest(t, g, me.ID, titanic, effectstest.PushCrewerForTest(g, me.ID, "Crewer", 3))

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(titanic, opp.ID, 6)
	})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(titanic) {
		t.Error("noncombat damage should not have sacrificed RMS Titanic")
	}
	if n := effectstest.B43TokensNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("Treasures = %d, want 0 (noncombat damage is not the trigger)", n)
	}
}
