package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dryad's Caress — Instant {4}{G}{G} (Ravnica: City of Guilds, S82 OLD b582):
//
//	"You gain 1 life for each creature on the battlefield. If {W} was spent to
//	 cast this spell, untap all creatures you control."
//
// Fog of War's life count over every player's creatures, then River's
// Grasp's reading of the colours the cast recorded (ReadsOwnPayment, so a
// waived payment asks what was paid and a copy has no mana spent). The untap
// is Aurelia's "untap all creatures you control", after the life (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "7df07d16-e5cf-4c58-980a-632909aa1357",
		Name:           "Dryad's Caress",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"W"},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Player: ctx.Controller(), Amount: len(ctx.CreatureIDs())}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if ctx.ManaSpentOfColor("W") == 0 {
					return nil
				}
				return effects.UntapAllCreaturesYouControl{}.Apply(ctx)
			}),
		),
	})
}
