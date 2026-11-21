package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Culling Dais — Artifact {2} (Scars of Mirrodin, #5577):
//
//	"{T}, Sacrifice a creature: Put a charge counter on this artifact.
//	 {1}, Sacrifice this artifact: Draw a card for each charge counter on
//	 this artifact."
//
// The first row is a sacrifice outlet whose counter goes on only while the
// Dais is still on the battlefield (b33PutChargeCounterOnSelf). The second
// sacrifices the Dais as its cost, so the count is read off the Dais as it
// last existed on the battlefield (CR 608.2h, Heliophial's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42b03dd7-e796-47c3-b596-87960d72818b",
		Name:         "Culling Dais",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Sacrifice a creature: Put a charge counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a creature", effects.Creature())),
				Effect:  effects.B33PutChargeCounterOnSelf,
			},
			{
				Label:   "{1}, Sacrifice this artifact: Draw a card for each charge counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					info, ok := ctx.SourcePermanent()
					if !ok || info.Counters[game.CounterCharge] <= 0 {
						return nil
					}
					return effects.DrawCards{N: info.Counters[game.CounterCharge]}.Apply(ctx)
				})),
			},
		},
	})
}
