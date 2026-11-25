package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illusionist's Stratagem — Instant {3}{U} (Aether Revolt, S82 older sets, #5536):
//
//	"Exile up to two target creatures you control, then return those cards
//	 to the battlefield under their owner's control.
//	 Draw a card."
//
// Displace's blink (every still-legal target is exiled at once and they
// return together, so each sees the other enter; a token does not come back,
// CR 111.8) with a draw that does not depend on a target (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1225bee2-c829-4a93-9f75-3495277c23bc",
		Name:         "Illusionist's Stratagem",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BragoBlinkTargets(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
