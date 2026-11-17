package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c0c4ecf6-84e6-443e-8a81-e395b31c631d",
		Name:         "Prizefight",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				other, ok := ctx.ClauseTarget(1)
				if !ok || other.Kind != game.TargetCard {
					return nil
				}
				return b10Fight(ctx, mine.ID, other.ID)
			}),
			Instruction(func(ctx *Context) error {
				return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
			}),
		),
	})
}
