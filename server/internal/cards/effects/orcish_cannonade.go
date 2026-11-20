package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orcish Cannonade — Instant {1}{R}{R} (Time Spiral, #5596):
//
//	"Orcish Cannonade deals 2 damage to any target and 3 damage to you.
//	 Draw a card."
//
// Char's body at 2 and 3 (one damage event, CR 120.3), then the draw. A
// target that became illegal counters the whole spell, so neither the
// self-damage nor the draw happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af224f6f-cedd-498f-a08e-7be0db5a11b7",
		Name:         "Orcish Cannonade",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove, Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := tmtB6DamageAnyTargetAndYou(2, 3)(ctx.Game, item); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
