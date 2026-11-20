package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blbHCastWithMana puts a card with a printed mana cost in the active
// player's hand, fills the pool with `pool` and casts it under strict
// mana, so a pool that can't pay the effective cost is refused.
func BlbHCastWithMana(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, power, toughness int, pool ...string) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	AdvanceToMain(t, g)
	B06AddMana(me, pool...)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
}
