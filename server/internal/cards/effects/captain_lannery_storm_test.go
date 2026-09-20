package effects

import (
	"testing"

	"github.com/google/uuid"
)

const captainLanneryStormOracle = "235bf0ba-658c-463f-b112-7478ba27bd7b"

// TestCaptainLanneryStormAttackMakesATreasureAndSacrificingItPumpsHer
// pins both triggers: the attack trigger creates a Treasure, and
// sacrificing that Treasure (the second trigger, watching
// EventSacrifice while the token is still on the battlefield) pumps
// her +1/+0 until end of turn.
func TestCaptainLanneryStormAttackMakesATreasureAndSacrificingItPumpsHer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lannery := pushDiesCreatureForTest(g, me.ID, "Captain Lannery Storm", captainLanneryStormOracle,
		"Legendary Creature — Human Pirate", 2, 2)

	declareAttack(t, g, opp.ID, lannery)
	passPriorityAroundTable(t, g)

	treasure := findBattlefieldByName(g, "Treasure")
	if treasure == uuid.Nil {
		t.Fatal("attacking creates a Treasure")
	}
	if p := effectivePower(t, g, lannery); p != 2 {
		t.Fatalf("no pump before any Treasure is sacrificed: power %d, want 2", p)
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(treasure) })
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, lannery); p != 3 {
		t.Errorf("sacrificing a Treasure pumps her +1/+0 until end of turn: power %d, want 3", p)
	}
}

// TestCaptainLanneryStormOnlyTreasuresTrigger the pump: sacrificing a
// non-Treasure permanent she also controls does nothing.
func TestCaptainLanneryStormOnlyTreasuresTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lannery := pushDiesCreatureForTest(g, me.ID, "Captain Lannery Storm", captainLanneryStormOracle,
		"Legendary Creature — Human Pirate", 2, 2)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, lannery); p != 2 {
		t.Errorf("sacrificing a non-Treasure must not pump her: power %d, want 2", p)
	}
}
