package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dead of Winter — Sorcery {2}{B} (Modern Horizons):
//
//	"All nonsnow creatures get -X/-X until end of turn, where X is the number
//	 of snow permanents you control."
//
// X is counted once, as the spell resolves, through the layered view (a
// permanent that is snow only by an effect counts), and the creatures it
// affects are fixed at that moment (CR 611.2c): a nonsnow creature that
// enters later this turn is untouched. Every nonsnow creature on the
// battlefield is affected, yours and theirs. With no snow permanents X is 0
// and nothing changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d877c28-5276-4ea6-91c6-eeed890ca545",
		Name:         "Dead of Winter",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := effects.S82OldB064SnowPermanentsYouControl(ctx.Game, ctx.Controller())
			if x <= 0 {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.Not(effects.Snow())),
				"Dead of Winter — nonsnow creatures get -X/-X", game.ModifyPTMod(-x, -x))
		},
	})
}
