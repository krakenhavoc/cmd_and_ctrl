package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hamletback Goliath — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "ae45b04e-c52d-45e1-a752-0b76f64d7c8d", Name: "Hamletback Goliath", Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Optional(On(game.EventETB, AnotherCreatureEntered, "Hamletback Goliath — counters equal to the entering creature's power", func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			info, ok := ctx.TriggeringPermanent()
			if !ok {
				return nil
			}
			return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: max(0, info.Power)}.Apply(ctx)
		}), "Hamletback Goliath — put counters on it?")},
	})
}
