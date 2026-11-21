package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castEntwineSpell puts `name` in the caster's hand and casts it with
// exactly `pool` in the mana pool, announcing `modes` and — when
// `entwined` — the card's entwine cost (its only optional cost). A
// refused cast takes the card back out of the hand and empties the pool,
// so the next attempt starts from the same board.
func CastEntwineSpell(g *game.Game, me *game.Player, name, typeLine, manaCost, oracle string, colors []string,
	pool string, modes []int, entwined bool, targets []game.TargetRef) (uuid.UUID, error) {
	id := HandCardFull(me, name, typeLine, manaCost, oracle, colors)
	me.ManaPool.EmptyPool()
	if pool != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, pool); err != nil {
			return id, err
		}
	}
	var optional []int
	if entwined {
		optional = []int{0}
	}
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, Modes: modes, OptionalCosts: optional, Targets: targets,
	})
	if err != nil {
		_, _ = me.Hand.Remove(id)
		me.ManaPool.EmptyPool()
	}
	return id, err
}
