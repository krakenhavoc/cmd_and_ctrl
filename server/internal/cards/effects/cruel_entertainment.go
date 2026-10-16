package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cruel Entertainment — Sorcery {6}{B}:
//
//	"Choose target player and another target player. The first player
//	 controls the second player during the second player's next turn,
//	 and the second player controls the first player during the first
//	 player's next turn."
//
// Two plain turn controls (CR 723.1, ADR 0167, #3857), one in each
// direction, with no extra turns. The caster only chooses the pair: they
// may be one of the two players or neither, and they control nobody
// unless they are the first or second player.
//
// Two target clauses (CR 601.2c), the second distinct from the first
// ("another"). Both players are named in both halves of the effect, so
// the spell does nothing if either has left the game by resolution
// (CR 608.2b: with an illegal target, no part of the effect that
// involves it happens); a player who has left is also refused by the
// engine.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7747804f-ca49-4a68-b508-b94b27c52b0f",
		Name:         "Cruel Entertainment",
		Completeness: CompletenessFull,
		Targets:      Clauses(TargetPlayer("target player"), Distinct(TargetPlayer("another target player"))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			first, ok1 := ctx.ClauseTarget(0)
			second, ok2 := ctx.ClauseTarget(1)
			if !ok1 || !ok2 {
				return nil
			}
			controlPlayerNextTurn(ctx.Game, item, first.ID, second.ID, "Cruel Entertainment")
			controlPlayerNextTurn(ctx.Game, item, second.ID, first.ID, "Cruel Entertainment")
			return nil
		},
	})
}
