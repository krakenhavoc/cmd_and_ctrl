package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const plainWalkerOracle = "c8d38d18-2669-494a-89a8-2ead0cc9628a"

func plainWalkerPush(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain Walker", TypeLine: "Creature — Kithkin Rogue", OracleID: plainWalkerOracle,
		Colors: []string{"W"}, Power: 2, Toughness: 3, Owner: owner, Controller: owner,
	})
}

func plainWalkerPlaneswalker(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Planeswalker — " + name,
		Counters: map[string]int{game.CounterLoyalty: 3}, Owner: owner, Controller: owner,
	})
}

func TestPlainWalkerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(plainWalkerOracle)
	if !ok || spec.Name != "Plain Walker" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Plain Walker spec = %+v, registered %v", spec, ok)
	}
}

func TestPlainWalkerCantBeBlockedWhileTheDefenderControlsAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := plainWalkerPush(g, me.ID)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	plainWalkerPlaneswalker(g, opp.ID, "Jace")
	effectstest.AdvanceToMain(t, g)
	effectstest.BrAttack(t, g, walker)
	if reason := effectstest.BlockRefusal(t, g, blocker, walker); reason != game.BlockReasonCantBeBlocked {
		t.Errorf("blocking with the defender's planeswalker out: reason %q, want %q", reason, game.BlockReasonCantBeBlocked)
	}
}

func TestPlainWalkerCanBeBlockedWithoutAPlaneswalkerAndOnlyTheDefendersCounts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	walker := plainWalkerPush(g, me.ID)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	// A planeswalker of a player who is not defending, and one of mine, do not count.
	plainWalkerPlaneswalker(g, third.ID, "Jace")
	plainWalkerPlaneswalker(g, me.ID, "Chandra")
	effectstest.AdvanceToMain(t, g)
	effectstest.BrAttack(t, g, walker)
	if reason := effectstest.BlockRefusal(t, g, blocker, walker); reason != "" {
		t.Errorf("the Walker was unblockable (%q) with no planeswalker on the defender's side", reason)
	}
}

func TestPlainWalkerPlainswalkIsStillEnforced(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := plainWalkerPush(g, me.ID)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	effectstest.AdvanceToMain(t, g)
	effectstest.BrAttack(t, g, walker)
	if reason := effectstest.BlockRefusal(t, g, blocker, walker); reason == "" {
		t.Error("the Walker was blocked although the defender controls a Plains")
	}
}

func TestPlainWalkerPlaneswalksWhenItDealsCombatDamageToAPlayer(t *testing.T) {
	g, up := effectstest.NewPlaneGame(t, effectstest.PlaneFixture{
		Oracle: "11111111-2222-4333-8444-555555555555", Name: "Plain Walker Test Plane", TypeLine: "Plane — Testworld"})
	me, opp := g.Seats[0], g.Seats[1]
	walker := plainWalkerPush(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, walker)
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("the Walker dealt %d damage, want 2", life-opp.Life)
	}
	effectstest.DrcSettle(t, g, me.ID)
	for _, c := range g.PlanarFaceUpCards() {
		if c.InstanceID == up.InstanceID {
			t.Error("the plane is still face up: no planeswalk happened")
		}
	}
}

func TestPlainWalkerOutsidePlanechaseJustDealsItsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := plainWalkerPush(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, walker)
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, opp.Life)
	}
}
