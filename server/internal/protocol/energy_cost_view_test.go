package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0129 §8: an activated row stamps its energy component, and a row
// the controller is short of energy for carries the engine's refusal as
// cant_activate, so the ready ring, the click and the engine agree.
func TestViewStampsEnergyCostAndShortfall(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	var src uuid.UUID
	g.WithWriteLock(func() {
		src = uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: src, Name: "Energy Rows", TypeLine: "Artifact",
			OracleID: "00000000-0000-0000-0000-000000001995",
			Owner:    me.ID, Controller: me.ID,
			ActivatedAbilities: []game.ActivatedAbilityShape{
				{Label: "fixed", Cost: game.AbilityCost{Energy: 3}},
				{Label: "x", Cost: game.AbilityCost{EnergyX: true}},
			},
		})
		if err := g.AddPlayerCounterForEffect(me.ID, game.CounterEnergy, 2); err != nil {
			t.Fatal(err)
		}
	})
	g.BumpLayerVersionForTest()

	rows := frameCard(t, g, me.ID.String(), src).ActivatedAbilities
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	if r := rows[0]; r.EnergyCost != 3 || r.EnergyCostX || r.CantActivate != "Not enough energy (have 2, need 3)" {
		t.Errorf("fixed row = energy %d x %v cant %q, want 3, false and the shortfall", r.EnergyCost, r.EnergyCostX, r.CantActivate)
	}
	if r := rows[1]; r.EnergyCost != 0 || !r.EnergyCostX || !r.DemandsX || r.CantActivate != "" {
		t.Errorf("x row = energy %d x %v demands_x %v cant %q, want 0, true, true and nothing", r.EnergyCost, r.EnergyCostX, r.DemandsX, r.CantActivate)
	}

	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(me.ID, game.CounterEnergy, 1); err != nil {
			t.Fatal(err)
		}
	})
	g.BumpLayerVersionForTest()
	if r := frameCard(t, g, me.ID.String(), src).ActivatedAbilities[0]; r.CantActivate != "" {
		t.Errorf("with 3 energy the row says %q, want nothing", r.CantActivate)
	}
}
