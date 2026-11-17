package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serum Snare — Instant {1}{U} (Phyrexia: All Will Be One, #5479):
//
//	"Return target nonland permanent to its owner's hand. If that permanent
//	 had mana value 3 or less, proliferate."
//
// The permanent's mana value is read before it moves (CR 608.2h). A
// bounced commander can pause for its owner's command-zone choice, so the
// proliferate is the bounce's keyed tail (thenProliferateAgain, one pass)
// and runs once the return is settled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbc7350c-d392-44bf-92f7-8dd4743a73b4",
		Name:         "Serum Snare",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			target := item.Targets[0].ID
			var then game.KeyedTail
			if s82OneB08ManaValueAtMost(ctx, target, 3) {
				then = ctx.TailN(thenProliferateAgain, 1)
			}
			return BounceToHand{Target: target, ThenTail: then}.Apply(ctx)
		},
	})
}
