package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2c42983e-f45a-4d10-afe5-d2234dfae1ee",
		Name:         "Overcooked",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		CantGainLife: PlayersCantGainLife(),
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Overcooked — create a Food token; with celebration, conjure a card named Food Fight onto the battlefield instead",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if Celebration(g, item.Controller) {
						return gapB05Conjure(ctx, item.Controller, gapB05FoodFightCard())
					}
					return CreateToken{Template: FoodToken(), N: 1}.Apply(ctx)
				}),
		},
	})
}
