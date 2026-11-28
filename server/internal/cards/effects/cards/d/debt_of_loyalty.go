package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Debt of Loyalty — Instant {1}{W}{W} (Weatherlight, S82 older sets batch 779, #5629):
//
//	"Regenerate target creature. You gain control of that creature if it regenerates this way."
//
// A regeneration shield with a rider (game.RegenerateGainingControlForEffect, CR 701.19a): if the shield this spell
// created is the one that replaces a destruction, the spell's controller gains control of the creature, with no
// duration. A creature that is never destroyed this turn is never taken; with a plain shield as well, the creature's
// controller spends the plain one first (CR 616.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a5cfc97-89cf-43bc-8e48-f8bf7a0c5fcd",
		Name:         "Debt of Loyalty",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil || ctx.IsNewSourceObject(id) {
				return nil
			}
			// CR 701.19b: regenerating a permanent that has left does nothing.
			_ = ctx.Game.RegenerateGainingControlForEffect(id, ctx.Controller(), item.SourceCardID, "Debt of Loyalty — gain control if it regenerates this way")
			return nil
		},
	})
}
