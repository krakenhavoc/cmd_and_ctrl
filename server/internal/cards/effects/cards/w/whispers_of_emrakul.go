package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whispers of Emrakul — Sorcery {1}{B} (Eldritch Moon, S82 OLD, #5541):
//
//	"Target opponent discards a card at random.
//	 Delirium — If there are four or more card types among cards in your
//	 graveyard, that player discards two cards at random instead."
//
// Flay's random discard with the count read as the spell resolves, from the
// caster's graveyard (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf8bc2e3-2f22-45d3-9810-bc039202f41c",
		Name:         "Whispers of Emrakul",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				n := 1
				if effects.DrcDeliriumMet(ctx.Game, ctx.Controller()) {
					n = 2
				}
				return effects.TargetPlayerDiscardsAtRandom{N: n}.Apply(ctx)
			}),
		),
	})
}
