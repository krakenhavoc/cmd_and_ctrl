package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Setessan Petitioner — Creature — Human Druid {1}{G}{G}, 2/2 (Theros Beyond
// Death, #5514):
//
//	"When this creature enters, you gain life equal to your devotion to green."
//
// Devotion is counted as the trigger resolves (CR 700.5), so the Petitioner's own
// {G}{G} is in it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63d92a82-d77d-4c47-a49b-bde6a8c9dd8d",
		Name:         "Setessan Petitioner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Setessan Petitioner — you gain life equal to your devotion to green",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GainLife{Player: item.Controller, Amount: effects.DevotionTo(g, item.Controller, ctx.TextColor("G"))}.Apply(ctx)
				}),
		},
	})
}
