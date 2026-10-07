package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_helpers_test.go — shared helpers for the ADR 0129 energy card
// tests (energy_*_test.go).

// energyOf is p's energy total.
func energyOf(p *game.Player) int { return game.PlayerEnergy(p) }

// setEnergy makes p's energy total exactly n.
func setEnergy(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddPlayerCounterForEffect(p.ID, game.CounterEnergy, n-game.PlayerEnergy(p)) })
	if err != nil {
		t.Fatalf("set energy: %v", err)
	}
}

// specActivatedEnergy is the Energy component of row `i` of a catalog
// card's own activated abilities, failing the test when it is not
// registered.
func specActivatedEnergy(t *testing.T, oracle string, i int) game.AbilityCost {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", oracle)
	}
	if i >= len(spec.Activated) {
		t.Fatalf("%s has %d activated rows, want row %d", spec.Name, len(spec.Activated), i)
	}
	return spec.Activated[i].Cost
}
