package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sorin_birthday_a_test.go — #4077 slice a: the four Sorin planeswalkers
// of the Sorin's Birthday Party deck request, driven through the real
// loyalty path with the real catalog wired.

const (
	SorinMirthlessOracle   = "fbffaf9f-40e7-4113-a037-533b733cebce"
	SorinGrimNemesisOracle = "0a01dd05-e289-489f-b3ad-88e15e157bd0"
	SorinLordOracle        = "5c780dda-ee76-482a-a610-53c321c4fdec"
	SorinSolemnOracle      = "1f1f1ae7-dd00-4825-832e-3a049cf8ce73"
)

// sorinATable is a fresh game at the active seat's precombat main phase
// with `name` on the battlefield at `loyalty`.
func SorinATable(t *testing.T, name, oracle string, loyalty int) (g *game.Game, seat int, me *game.Player, sorin uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	seat = g.Turn.ActiveSeat
	me = g.Seats[seat]
	sorin = PushWalkerForTest(g, me.ID, name, oracle, loyalty)
	AdvanceToMainOf(t, g, seat)
	return g, seat, me, sorin
}
