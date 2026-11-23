package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hamletback Goliath — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "ae45b04e-c52d-45e1-a752-0b76f64d7c8d", Name: "Hamletback Goliath", Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Optional(effects.On(game.EventETB, effects.AnotherCreatureEntered, "Hamletback Goliath — counters equal to the entering creature's power", func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			info, ok := ctx.TriggeringPermanent()
			if !ok {
				return nil
			}
			return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: max(0, info.Power)}.Apply(ctx)
		}), "Hamletback Goliath — put counters on it?")},
	})
}
