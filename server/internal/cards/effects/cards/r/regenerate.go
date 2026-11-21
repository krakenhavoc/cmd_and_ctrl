package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Regenerate — Instant {1}{G} (Magic 2010, #5585):
//
//	"Regenerate target creature. (The next time that creature would be
//	 destroyed this turn, instead tap it, remove it from combat, and heal
//	 all damage on it.)"
//
// Refresh without the draw: one regeneration shield on the target
// (CR 701.19a). If the target is gone, the spell does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ce6c946-7a93-4a8d-a2c9-086cbd036e3c",
		Name:         "Regenerate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.RegenerateTheTargetPermanent(ctx.Game, item)
		},
	})
}
