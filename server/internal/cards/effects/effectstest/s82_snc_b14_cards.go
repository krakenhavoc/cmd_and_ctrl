package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sncB14Cast puts the card in `me`'s hand, floats `pool` (one letter per mana,
// "C" for colourless) and casts it under the strict mana gate.
func SncB14Cast(t *testing.T, g *game.Game, me *game.Player, card game.Card, pool string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	// An instant cast during combat stays in combat: only a turn still
	// before its first main phase is walked forward.
	if g.Turn.Step != game.StepDeclareAttackers && g.Turn.Step != game.StepDeclareBlockers {
		ToMainForCost(t, g)
	}
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = me.ID, me.ID
	me.Hand.PushTop(card)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	FloatForTest(g, me, pool)
	params.Strict = true
	return card.InstanceID, g.CastSpell(me.ID, card.InstanceID, params)
}
