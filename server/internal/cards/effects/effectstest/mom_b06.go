package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func MomCast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	params.Strict = true
	return CastFromHandForTest(t, g, me, name, typeLine, cost, oracle, params)
}

func MomCastErr(g *game.Game, me *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) error {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: oracle, Owner: me.ID, Controller: me.ID})
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}
