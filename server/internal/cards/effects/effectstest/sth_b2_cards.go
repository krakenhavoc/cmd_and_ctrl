package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sthB2Spike seeds a Spike with n +1/+1 counters on the active seat's
// battlefield, in its main phase.
func SthB2Spike(t *testing.T, g *game.Game, name, typeLine, oracle string, n int) (*game.Player, uuid.UUID) {
	t.Helper()
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := TmpB1Push(g, me.ID, name, typeLine, oracle, 0, 0, "G")
	if n > 0 {
		g.WithWriteLock(func() { FindBattlefieldCardForTest(g, id).Counters = map[string]int{game.CounterPlusOne: n} })
	}
	return me, id
}
