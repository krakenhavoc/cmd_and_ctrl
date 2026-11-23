package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b11Cast casts a creature with the given catalog identity from the
// active seat's hand in a main phase and resolves it.
func B11Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	AdvanceToMain(t, g)
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID}
	CastFromHandAndResolve(t, g, me, c)
	DrcSettle(t, g, me.ID)
	return c.InstanceID
}

func B11Die(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	KillCreature(t, g, owner, id)
	DrcSettle(t, g, owner)
}
