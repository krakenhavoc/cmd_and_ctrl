package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// energy_test.go — ADR 0129 §7: an activation that pays energy is
// offered only when the seat has the energy, its MoveCost names the
// energy, "Pay X {E}" is bounded by the seat's energy, and every move
// offered is one the dispatcher accepts (#544).

func giveEnergy(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddPlayerCounterForEffect(p.ID, game.CounterEnergy, n) })
	if err != nil {
		t.Fatalf("give energy: %v", err)
	}
}

func TestEnergyCostIsGatedOnTheSeatsEnergy(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	src := battlefieldCard(g, active, lcSource(g, active, game.AbilityCost{Energy: 3}))
	advanceTo(t, g, game.StepPrecombatMain)

	giveEnergy(t, g, active, 2)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), src); len(acts) != 0 {
		t.Fatalf("two energy: want no move, got %v", labels(acts))
	}
	giveEnergy(t, g, active, 1)
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, src)
	if len(acts) != 1 {
		t.Fatalf("three energy: want one move, got %v", labels(acts))
	}
	if c := acts[0].Cost; c == nil || c.Energy != 3 {
		t.Errorf("cost = %+v, want energy 3", c)
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestEnergyXIsBoundedByTheSeatsEnergy(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	src := battlefieldCard(g, active, lcSource(g, active, game.AbilityCost{EnergyX: true}))
	advanceTo(t, g, game.StepPrecombatMain)

	giveEnergy(t, g, active, 5)
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, src)
	if len(acts) != 1 {
		t.Fatalf("want one move, got %v", labels(acts))
	}
	var p struct {
		XValue int `json:"x_value"`
	}
	if err := json.Unmarshal(acts[0].Params, &p); err != nil {
		t.Fatal(err)
	}
	if p.XValue != 5 {
		t.Errorf("x_value = %d, want 5 (the seat's energy)", p.XValue)
	}
	if c := acts[0].Cost; c == nil || c.Energy != 5 {
		t.Errorf("cost = %+v, want energy 5", c)
	}
	if v := acts[0].Value; v == nil || v.Max == nil || *v.Max != 5 {
		t.Errorf("value = %+v, want an open X up to 5", v)
	}
	dispatchAll(t, g, active.ID, moves)
}
