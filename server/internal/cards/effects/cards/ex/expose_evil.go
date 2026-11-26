package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Expose Evil — Instant {1}{W} (Shadows over Innistrad, #5543):
//
//	"Tap up to two target creatures.
//	 Investigate."
//
// Up to two targets, each tapped if still legal as the spell resolves
// (CR 608.2b). With no target chosen the spell still resolves: the Clue is
// made, and it is made even if every target has become illegal only when
// none was chosen (a spell whose chosen targets are all illegal does
// nothing, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5483037-5dd2-4282-a201-de2e08aaa76e",
		Name:         "Expose Evil",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.TapEachLegalTarget(ctx.Game, ctx.Item) }),
			effects.Investigate{N: 1},
		),
	})
}
