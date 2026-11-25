package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temporal Machinations — Sorcery {2}{U} (Dominaria, S82 OLD, #5530):
//
//	"Return target creature to its owner's hand. If you control an artifact,
//	 draw a card."
//
// The artifact is read as the spell resolves, after the bounce. A creature
// that became an illegal target fizzles the whole spell (CR 608.2b), so
// nothing is drawn then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0adab66-c3fe-4989-b7c8-ab7bdc7f0fd1",
		Name:         "Temporal Machinations",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				ts := ctx.LegalTargets()
				if len(ts) == 0 || ts[0].Kind != game.TargetCard || ts[0].ID == uuid.Nil {
					return nil
				}
				return effects.BounceToHand{Target: ts[0].ID}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Artifact())) == 0 {
					return nil
				}
				return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
