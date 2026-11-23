package testsm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const miserysShadowOracle = "928c62fe-9c4d-4e89-bd7a-0d3b6a81f393"

func TestMiserysShadowIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(miserysShadowOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
}

// An opponent's creature that would die is exiled; their token too; the
// controller's own creature dies normally.
func TestMiserysShadowExilesOpponentCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Misery's Shadow", "Creature — Shade", miserysShadowOracle, false)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
		if err := g.DestroyPermanentForEffect(mine); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !g.Exile.Contains(bear) || opp.Graveyard.Contains(bear) {
		t.Error("the opponent's creature should be exiled, not in the graveyard")
	}
	if !me.Graveyard.Contains(mine) || g.Exile.Contains(mine) {
		t.Error("the controller's own creature should die normally")
	}
}

// Lethal damage (a state-based action) is exiled too.
func TestMiserysShadowOnAStateBasedDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Misery's Shadow", "Creature — Shade", miserysShadowOracle, false)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := g.MarkDamage(bear, 2); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature that died to damage was not exiled")
	}
}

// {1}: +1/+1, repeatable, ending at end of turn.
func TestMiserysShadowPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	src := effectstest.PushCatalogPermanent(g, me.ID, "Misery's Shadow", "Creature — Shade", miserysShadowOracle, true)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, src)
		c.Power, c.Toughness = 2, 2
	})
	for range 2 {
		effectstest.ApaMana(me, "C")
		effectstest.ApaActivate(t, g, me, src, 0, game.ActivateAbilityParams{})
	}
	if p, tt := effectstest.PtOf(t, g, src); p != 4 || tt != 4 {
		t.Errorf("after two pumps %d/%d, want 4/4", p, tt)
	}
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, tt := effectstest.PtOf(t, g, src); p != 2 || tt != 2 {
		t.Errorf("next turn %d/%d, want 2/2", p, tt)
	}
}
