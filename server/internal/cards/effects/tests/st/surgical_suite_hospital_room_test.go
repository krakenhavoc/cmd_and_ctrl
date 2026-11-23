package testsst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const surgicalSuiteOracle = "68bb76af-f933-430a-a66a-fe4a7d6be4e0"

// Surgical Suite offers only a creature card with mana value 3 or less
// from your graveyard; Hospital Room then counts once per attack, on
// the attacking creature named when the trigger goes on the stack, and
// not while it is locked.
func TestSurgicalSuiteHospitalRoomDoors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	small := effectstest.PushGraveyardPermanent(me, "Small Dead", "Creature — Bear", "{1}{W}")
	big := effectstest.PushGraveyardPermanent(me, "Big Dead", "Creature — Giant", "{4}{W}")
	c := effectstest.RoomsBCard(me.ID, surgicalSuiteOracle, "Surgical Suite", "{1}{W}", "Hospital Room", "{3}{W}")
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)

	effectstest.RoomsBCast(t, g, me, c, 0)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me.ID) == nil && !effectstest.StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("Surgical Suite asked for no target")
	}
	if !effectstest.HasID(p.PickTargetCards, small) || effectstest.HasID(p.PickTargetCards, big) {
		t.Fatalf("offered %v, want only the mana value 2 creature card", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.RoomsBSettle(t, g, me)
	if !effects.OnBattlefield(g, small) || !me.Graveyard.Contains(big) {
		t.Fatal("Surgical Suite did not reanimate exactly the chosen card")
	}

	// Hospital Room is locked: attacking puts no counter anywhere.
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.RoomsBSettle(t, g, me)
	if n := effectstest.FindBattlefieldCardForTest(g, attacker).Counters[game.CounterPlusOne]; n != 0 {
		t.Fatalf("a locked Hospital Room put %d counters", n)
	}
}

func TestHospitalRoomCountsOncePerAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	c := effectstest.RoomsBCard(me.ID, surgicalSuiteOracle, "Surgical Suite", "{1}{W}", "Hospital Room", "{3}{W}")
	a := effectstest.PushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Attacker B", 2, 2)

	effectstest.RoomsBCast(t, g, me, c, 1)
	effectstest.RoomsBSettle(t, g, me)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.RoomsBSettle(t, g, me, b)
	ca := effectstest.FindBattlefieldCardForTest(g, a).Counters[game.CounterPlusOne]
	cb := effectstest.FindBattlefieldCardForTest(g, b).Counters[game.CounterPlusOne]
	if ca != 0 || cb != 1 {
		t.Fatalf("counters a=%d b=%d, want 0 and 1 (one trigger for two attackers, on the chosen one)", ca, cb)
	}
}
