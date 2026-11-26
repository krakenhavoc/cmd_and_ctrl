package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaotic Backlash — Instant {4}{R} (Eventide, S82 OLD batch 520, #5590):
//
//	"Chaotic Backlash deals damage to target player equal to twice the number of
//	 white and/or blue permanents they control."
//
// The count is read as the spell resolves, through the layers: each permanent
// the target player controls that is white, blue or both counts once (a
// white-and-blue one is not counted twice), and twice that is the damage. A
// target gone by resolution makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be6cb3e6-7492-4e30-83a9-9bc413e50cdc",
		Name:         "Chaotic Backlash",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			n := len(effects.PermanentsControlledByMatching(ctx.Game, player, effects.Or(effects.OfColor("W"), effects.OfColor("U"))))
			return effects.DealDamage{Source: ctx.Source(), Target: player, Amount: 2 * n}.Apply(ctx)
		},
	})
}
