package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdnCCast casts a catalog creature from the active seat's hand in a main
// phase and resolves it and its entry trigger.
func FdnCCast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID}
	CastFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}
