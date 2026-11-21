package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overcooked — Enchantment {2}{R} (Mystery Booster Commander, #5948):
//
//	"Players can't gain life.
//	 Celebration — At the beginning of your end step, create a Food token. If
//	 two or more nonland permanents entered the battlefield under your control
//	 this turn, instead conjure a card named Food Fight onto the battlefield."
//
// Players can't gain life is Forsaken Wastes' static (CR 119.7), and it
// stops Food's gain too. The end-step trigger is not an intervening if: the
// celebration is read as it resolves (Celebration, CR 207.2c), and the
// Food is created or the Food Fight conjured, never both. The conjured card
// is a real Food Fight (gapB05Conjure) that enters under your control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c42983e-f45a-4d10-afe5-d2234dfae1ee",
		Name:         "Overcooked",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		CantGainLife: effects.PlayersCantGainLife(),
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Overcooked — create a Food token; with celebration, conjure a card named Food Fight onto the battlefield instead",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.Celebration(g, item.Controller) {
						return effects.GapB05Conjure(ctx, item.Controller, effects.GapB05FoodFightCard())
					}
					return effects.CreateToken{Template: effects.FoodToken(), N: 1}.Apply(ctx)
				}),
		},
	})
}
