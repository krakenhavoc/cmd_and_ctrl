package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// speedGame is a catalog game walked to the active seat's precombat
// main (every max-speed sorcery ability is legal there, and a manual
// draw is not swallowed by the turn's draw step).
func SpeedGame(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceTo(t, g, game.StepPrecombatMain)
	return g, g.Seats[g.Turn.ActiveSeat]
}

// pushSpeedCard puts a catalog card on the battlefield as a non-sick
// permanent and runs the state-based actions, so its start your
// engines! has given its controller speed.
func PushSpeedCard(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner.ID, Controller: owner.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	g.RunStateChecksForTest()
	return id
}
