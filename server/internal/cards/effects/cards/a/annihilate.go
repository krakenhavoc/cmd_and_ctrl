package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Annihilate — Instant {3}{B}{B} (Eternal Masters, S82 OLD, #5542):
//
//	"Destroy target nonblack creature. It can't be regenerated.
//	 Draw a card."
//
// Doom Blade's target clause with Befoul's no-regeneration destroy, then the
// draw. The target is re-checked as the spell resolves (CR 608.2b): a creature
// that turned black in response is illegal, and with no legal target nothing
// happens, the draw included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28f55673-7aa0-497e-8434-aa70df1b5e04",
		Name:         "Annihilate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyTheTargetPermanentNoRegen(ctx.Item, ctx) }),
			effects.DrawCards{N: 1},
		),
	})
}
