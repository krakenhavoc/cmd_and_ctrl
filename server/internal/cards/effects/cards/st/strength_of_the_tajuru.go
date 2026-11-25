package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strength of the Tajuru — Instant {X}{G}{G} (Worldwake; OLD b121, #6150):
//
//	"Multikicker {1} (You may pay an additional {1} any number of times
//	 as you cast this spell.)
//	 Choose target creature, then choose another target creature for
//	 each time this spell was kicked. Put X +1/+1 counters on each of
//	 them."
//
// The clause is counted by the kicks (CountedByKicks, CR 601.2c): one
// creature unkicked, one more for each payment of {1}, all different.
// Each creature still a legal target as it resolves gets X +1/+1
// counters (CR 608.2b). The multikicker cap of 20 is the engine's bound
// on one announcement.
//
// No simplification.
func init() {
	targets := effects.CountedByKicks(effects.TargetCreature("target creature, and another for each time this spell was kicked"))
	effects.Register(effects.Spec{
		OracleID:      "132401fb-6754-4190-9c85-1e57e7f63ed6",
		Name:          "Strength of the Tajuru",
		Completeness:  effects.CompletenessFull,
		XMatters:      true,
		OptionalCosts: []game.AdditionalCost{effects.Multikicker("{1}", 20)},
		Targets:       targets,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: x}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
