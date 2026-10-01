package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const surgicalSuiteOracle = "68bb76af-f933-430a-a66a-fe4a7d6be4e0"

// Surgical Suite offers only a creature card with mana value 3 or less
// from your graveyard; Hospital Room then counts once per attack, on
// the attacking creature named when the trigger goes on the stack, and
// not while it is locked.
func TestSurgicalSuiteHospitalRoomDoors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	small := pushGraveyardPermanent(me, "Small Dead", "Creature — Bear", "{1}{W}")
	big := pushGraveyardPermanent(me, "Big Dead", "Creature — Giant", "{4}{W}")
	c := roomsBCard(me.ID, surgicalSuiteOracle, "Surgical Suite", "{1}{W}", "Hospital Room", "{3}{W}")
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)

	roomsBCast(t, g, me, c, 0)
	advanceToMain(t, g)
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("Surgical Suite asked for no target")
	}
	if !hasID(p.PickTargetCards, small) || hasID(p.PickTargetCards, big) {
		t.Fatalf("offered %v, want only the mana value 2 creature card", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, small)
	roomsBSettle(t, g, me)
	if !onBattlefield(g, small) || !me.Graveyard.Contains(big) {
		t.Fatal("Surgical Suite did not reanimate exactly the chosen card")
	}

	// Hospital Room is locked: attacking puts no counter anywhere.
	declareAttack(t, g, opp.ID, attacker)
	roomsBSettle(t, g, me)
	if n := findBattlefieldCardForTest(g, attacker).Counters[game.CounterPlusOne]; n != 0 {
		t.Fatalf("a locked Hospital Room put %d counters", n)
	}
}

func TestHospitalRoomCountsOncePerAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	c := roomsBCard(me.ID, surgicalSuiteOracle, "Surgical Suite", "{1}{W}", "Hospital Room", "{3}{W}")
	a := pushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Attacker B", 2, 2)

	roomsBCast(t, g, me, c, 1)
	roomsBSettle(t, g, me)
	declareAttack(t, g, opp.ID, a, b)
	roomsBSettle(t, g, me, b)
	ca := findBattlefieldCardForTest(g, a).Counters[game.CounterPlusOne]
	cb := findBattlefieldCardForTest(g, b).Counters[game.CounterPlusOne]
	if ca != 0 || cb != 1 {
		t.Fatalf("counters a=%d b=%d, want 0 and 1 (one trigger for two attackers, on the chosen one)", ca, cb)
	}
}
