package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b381_helpers.go — shared bodies for Commander 2013 batch 381
// (#5560).

// OldB381CopyTargetSpellTimes is "Copy target instant or sorcery spell N
// times. You may choose new targets for the copies" for the first target
// still legal as the ability resolves (CR 707.10c): Echo Mage's two bands.
func OldB381CopyTargetSpellTimes(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		targets := ctx.LegalTargets()
		if len(targets) == 0 {
			return nil
		}
		for i := 0; i < n; i++ {
			if err := (CopySpell{StackID: targets[0].ID, Controller: item.Controller, ChooseNewTargets: true}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// OldB381PlusOneOnEachCreature is "put a +1/+1 counter on each creature
// <player> controls": the set is read once, before the first counter lands.
func OldB381PlusOneOnEachCreature(ctx *Context, player uuid.UUID) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	return B13PutCounterOnEach(ctx, ids)
}

// OldB381LookAtTopXTakeOne is "Look at the top X cards of your library.
// Put one of them into your hand and the rest on the bottom of your library
// in any order" with X read off the effect.
func OldB381LookAtTopXTakeOne(x int, label string) Effect {
	return LookAtTopThenTakeOneRestOnBottom(x, label)
}
