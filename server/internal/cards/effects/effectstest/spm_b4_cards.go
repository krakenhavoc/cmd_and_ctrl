package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spmB4Cast puts a card with real P/T in the active player's hand and casts
// it for free (X given), returning its id and the cast's error.
func SpmB4Cast(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, power, toughness, x int) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{XValue: x})
}
