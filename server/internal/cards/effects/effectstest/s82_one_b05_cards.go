package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func CastCatalogSpellErrStrict(t *testing.T, g *game.Game, name, typeLine, oracleID string, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID})
	return g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets, Strict: true})
}
