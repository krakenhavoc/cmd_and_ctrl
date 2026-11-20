package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "42b03dd7-e796-47c3-b596-87960d72818b",
		Name:         "Culling Dais",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}, Sacrifice a creature: Put a charge counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    Plus(TapCost(), SacrificeN(1, "a creature", Creature())),
				Effect:  b33PutChargeCounterOnSelf,
			},
			{
				Label:   "{1}, Sacrifice this artifact: Draw a card for each charge counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
				Effect: Do(Instruction(func(ctx *Context) error {
					info, ok := ctx.SourcePermanent()
					if !ok || info.Counters[game.CounterCharge] <= 0 {
						return nil
					}
					return DrawCards{N: info.Counters[game.CounterCharge]}.Apply(ctx)
				})),
			},
		},
	})
}
