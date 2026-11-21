package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cosmos Elixir — Artifact {4} (EDHREC rank 2719):
//
//	"At the beginning of your end step, draw a card if your life total
//	 is greater than your starting life total. Otherwise, you gain 2
//	 life."
//
// The lifegain deck's card-advantage engine: two life a turn until
// the total is above 40, then a card a turn for as long as it stays
// there. One end-step trigger ("your" end step — ev.Actor is the
// active player) that reads the controller's life at RESOLUTION
// against the format's starting total (b25LifeAboveStarting), so a
// drain in response to the trigger turns the draw back into the
// lifegain, as printed. The life goes through the effect path, so
// every lifegain payoff sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed7300f4-831a-4ba4-b5e6-ceba8d079eaa",
		Name:         "Cosmos Elixir",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Cosmos Elixir — draw a card if above starting life, otherwise gain 2 life", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.B25LifeAboveStarting(g, item.Controller) {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				}
				return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
			}),
		},
	})
}
