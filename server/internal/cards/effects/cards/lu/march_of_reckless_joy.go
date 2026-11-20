package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// March of Reckless Joy — Instant {X}{R} (Kamigawa: Neon Dynasty):
//
//	"As an additional cost to cast this spell, you may exile any number
//	 of red cards from your hand. This spell costs {2} less to cast for
//	 each card exiled this way.
//	 Exile the top X cards of your library. You may play up to two of
//	 those cards until the end of your next turn."
//
// The cost is March of Swirling Mist's. The cards that land are ONE
// permission (CR 611.2c locks the set) with a budget of two
// (CastPermission.CastsLeft): a cast or a land play spends one, and the
// grant ends when the second is spent or at the end of the controller's next
// turn. A land played this way still uses a land drop (CR 305.2). X = 0
// exiles nothing and grants nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "3b3ecaf2-7e95-4b62-9bc0-a61985508808",
		Name:           "March of Reckless Joy",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.ExileAnyNumberFromHandCost("red cards", "R"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessPerExiledFromHand("{2}", "This spell costs {2} less to cast for each card exiled this way"),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.X() <= 0 {
				return nil
			}
			controller := ctx.Controller()
			_, err := ctx.Game.ExileTopWithPermissionForEffect(controller, controller, ctx.X(), game.CastPermission{
				Duration:   ctx.Game.UntilEndOfYourNextTurnDuration(controller),
				CastsLeft:  2,
				Source:     ctx.Source(),
				SourceName: "March of Reckless Joy",
				Label:      "Play up to two of the exiled cards until the end of your next turn",
			})
			return err
		},
	})
}
