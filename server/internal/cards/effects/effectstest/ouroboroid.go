package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const OuroboroidOracle = "50d6fd91-23d3-4d32-804f-6233e4386904"

// ouroboroidCreature pushes a vanilla creature with printed P/T. The
// #2401 board's cards stand in as these: their own text plays no part
// in the counters.
func OuroboroidCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// ouroboroidSetCounters sets a battlefield card's +1/+1 counters
// directly, standing in for counters it got earlier in the game.
func OuroboroidSetCounters(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: n}
			}
		}
	})
}

// ouroboroidToCombat walks to seat 0's next beginning of combat (it
// always leaves the current step first) without resolving anything.
func OuroboroidToCombat(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepBeginCombat {
			return
		}
	}
	t.Fatal("never reached seat 0's beginning of combat")
}
