package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Touch of the Eternal — Enchantment {5}{W}{W} (Magic 2013, S82 OLD, #5568):
//
//	"At the beginning of your upkeep, count the number of permanents you
//	 control. Your life total becomes that number."
//
// The count is taken as the trigger resolves (CR 608.2h) over every permanent
// you control, tokens and lands included, this enchantment too. The life change
// is Oketra's Last Mercy's (CR 119.5): the difference is gained or lost, so
// lifegain and life-loss triggers see it and "can't gain life" stops the gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23ecb4bb-b295-46f5-9fd8-a4c66ce80c23",
		Name:         "Touch of the Eternal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Touch of the Eternal — your life total becomes the number of permanents you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.B31LifeBecomes(ctx, ctx.Controller(), effects.S82OldB429PermanentsYouControl(ctx))
				}),
		},
	})
}
