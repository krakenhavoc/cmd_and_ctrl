package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Remand — Instant {1}{U} (Foundations Jumpstart):
//
//	"Counter target spell. If that spell is countered this way, put it
//	 into its owner's hand instead of into that player's graveyard.
//	 Draw a card."
//
// A real counter with a destination (CR 701.6a): the spell goes to its
// OWNER's hand, so a spell that can't be countered stays on the stack and
// is neither countered nor returned. The draw is the spell's own and
// happens regardless of whether the counter did anything, so long as the
// target was still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d71cd08e-3e84-41ff-b9db-9e343c0af6b4",
		Name:         "Remand",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				item := ctx.Game.StackItemForEffect(t.ID)
				if item == nil {
					return nil
				}
				return effects.CounterTarget{StackID: t.ID, Dest: game.ZoneRef{Kind: game.ZoneHand, Owner: item.Owner}}.Apply(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
