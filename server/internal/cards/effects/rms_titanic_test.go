package effects

import "testing"

const rmsTitanicOracle = "6754753d-790e-4a95-a150-76741c4a02d4"

// TestRMSTitanicSacrificesAndMakesThatManyTreasures — "that many" is
// the damage actually dealt, and the sacrifice and the tokens both
// happen off the one trigger.
func TestRMSTitanicSacrificesAndMakesThatManyTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	titanic := pushVehicleForTest(g, me.ID, "RMS Titanic", rmsTitanicOracle, 6, 6)
	crewForTest(t, g, me.ID, titanic, pushCrewerForTest(g, me.ID, "Crewer", 3))
	assertKeywords(t, g, titanic, "flying", "trample")

	dealCombatDamageToPlayer(g, titanic, opp.ID, 6)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(titanic) {
		t.Error("RMS Titanic should have sacrificed itself")
	}
	if n := b43TokensNamed(g, me.ID, "Treasure"); n != 6 {
		t.Fatalf("Treasures = %d, want 6 (the damage dealt)", n)
	}
}

// TestRMSTitanicNoncombatDamageIsNotTheTrigger — the ability is
// explicitly about COMBAT damage; a noncombat hit for the same amount
// must not sacrifice it or make any Treasures.
func TestRMSTitanicNoncombatDamageIsNotTheTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	titanic := pushVehicleForTest(g, me.ID, "RMS Titanic", rmsTitanicOracle, 6, 6)
	crewForTest(t, g, me.ID, titanic, pushCrewerForTest(g, me.ID, "Crewer", 3))

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(titanic, opp.ID, 6)
	})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(titanic) {
		t.Error("noncombat damage should not have sacrificed RMS Titanic")
	}
	if n := b43TokensNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("Treasures = %d, want 0 (noncombat damage is not the trigger)", n)
	}
}
