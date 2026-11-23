package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deep Space Scan — Sorcery {4}{U}{U} (Star Trek, #4973):
//
//	"Strange new worlds — Draw X cards, where X is the number of
//	 differently named lands you control. You have no maximum hand size
//	 until your next turn."
//
// "Strange new worlds" is an ability word with no rules meaning. X is
// counted as the spell resolves (each land name once), and the hand-size
// change is a resolution grant (Enter the Infinite's), so a countered Scan
// leaves the maximum alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5dabb7c1-8937-42b0-87c6-1cc7af46aaae",
		Name:         "Deep Space Scan",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DrawCards{Player: item.Controller, N: effects.SocB23DifferentlyNamedLands(ctx)}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.GrantHandSizeForEffect(item.Controller, game.HandSizeNoMaximum, 0,
				"Deep Space Scan — no maximum hand size", item.SourceCardID,
				effects.DurationUntilYourNextTurn(ctx, item.Controller))
		},
	})
}
