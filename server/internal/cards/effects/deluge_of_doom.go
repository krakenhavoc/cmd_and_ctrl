package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deluge of Doom — Sorcery {2}{B} (Duskmourn Commander):
//
//	"All creatures get -X/-X until end of turn, where X is the number of card
//	 types among cards in your graveyard."
//
// X is counted once as the spell resolves (the spell is still on the stack,
// so it is not counted) and the set of creatures is locked as the effect
// begins (CR 611.2c), as BoostUntilEOT does for every "all creatures" boost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a81fcdf8-59e7-471d-96ce-f87e656572b1",
		Name:         "Deluge of Doom",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := b16CardTypesInGraveyard(ctx.Game, item.Controller)
			if x == 0 {
				return nil
			}
			return BoostUntilEOT{Match: Creature(), Power: -x, Toughness: -x,
				Label: "Deluge of Doom — all creatures get -X/-X"}.Apply(ctx)
		},
	})
}
