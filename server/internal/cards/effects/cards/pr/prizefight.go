package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prizefight — Instant {1}{G} (Streets of New Capenna, #5491):
//
//	"Target creature you control fights target creature you don't control.
//	 Create a Treasure token."
//
// The Treasure is made even when the fight does not happen because one
// fighter is gone (CR 701.14b), as long as the spell still resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0c4ecf6-84e6-443e-8a81-e395b31c631d",
		Name:         "Prizefight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				other, ok := ctx.ClauseTarget(1)
				if !ok || other.Kind != game.TargetCard {
					return nil
				}
				return effects.B10Fight(ctx, mine.ID, other.ID)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
			}),
		),
	})
}
