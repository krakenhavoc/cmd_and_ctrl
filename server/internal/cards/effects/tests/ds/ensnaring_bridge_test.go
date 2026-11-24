package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ensnaringBridgeOracle = "dade5835-ccaf-45e3-b9a3-46017e0384b3"

func TestEnsnaringBridgeIsFull(t *testing.T) {
	spec, ok := effects.Lookup(ensnaringBridgeOracle)
	if !ok || spec.Name != "Ensnaring Bridge" {
		t.Fatalf("registered=%v name=%q", ok, spec.Name)
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%v %q, want Full with no caveat", spec.Completeness, spec.Caveats)
	}
}

// bridgeFillHand leaves exactly n filler cards in p's hand.
func bridgeFillHand(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for p.Hand.Size() > 0 {
			p.Hand.Remove(p.Hand.Cards[0].InstanceID)
		}
		for i := 0; i < n; i++ {
			p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
		}
	})
}

func pushEnsnaringBridge(g *game.Game, controller uuid.UUID) uuid.UUID {
	return effectstest.PushTestPermanent(g, controller, game.Card{Name: "Ensnaring Bridge", OracleID: ensnaringBridgeOracle, TypeLine: "Artifact"})
}

// An opponent's Bridge: their hand is the bound, power above it can't
// attack, power equal to it can, and a card they draw changes the answer.
func TestEnsnaringBridgeBindsPowerAboveItsControllersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushEnsnaringBridge(g, them.ID)
	big := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	even := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	bridgeFillHand(g, me, 7) // the attacker's own hand is not the Bridge's
	bridgeFillHand(g, them, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)

	effectstest.WantAttackRefused(t, g, big, them.ID, "greater than the number of cards in")
	if err := effectstest.DeclareResult(g, even, them.ID); err != nil {
		t.Errorf("power equal to the hand size: %v", err)
	}
	bridgeFillHand(g, them, 3)
	if err := effectstest.DeclareResult(g, big, them.ID); err != nil {
		t.Errorf("power 3 with three cards in the Bridge controller's hand: %v", err)
	}
	bridgeFillHand(g, them, 0)
	effectstest.WantAttackRefused(t, g, even, them.ID, "Ensnaring Bridge")
}

// Your own Bridge binds your own creatures, by your hand; power is read
// through the layers; and the restriction leaves with the Bridge.
func TestEnsnaringBridgeBindsItsControllersCreaturesAndLeavesWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bridge := pushEnsnaringBridge(g, me.ID)
	elf := effectstest.B12Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	bridgeFillHand(g, me, 1)
	bridgeFillHand(g, them, 7)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := effectstest.DeclareResult(g, elf, them.ID); err != nil {
		t.Fatalf("power 1, hand 1: %v", err)
	}
	effectstest.PutCounters(t, g, elf, game.CounterPlusOne, 1)
	effectstest.WantAttackRefused(t, g, elf, them.ID, "its power (2)")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bridge); err != nil {
		t.Fatalf("move the Bridge: %v", err)
	}
	if err := effectstest.DeclareResult(g, elf, them.ID); err != nil {
		t.Errorf("with the Bridge gone: %v", err)
	}
}
