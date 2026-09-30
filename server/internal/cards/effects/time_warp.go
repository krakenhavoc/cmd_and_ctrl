package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Time Warp — Sorcery {3}{U}{U}:
//
//	"Target player takes an extra turn after this one."
//
// CR 500.7 through TakeExtraTurn (ADR 0059 Decision 5, #753). "Target
// player" is any player, so it can give an opponent the turn — a
// legal, if odd, play. The turn is queued as the spell resolves and
// begins once this one ends; the most recently created extra turn is
// taken first, so a second Time Warp cast during the extra turn comes
// before anything queued earlier. The target is re-checked at
// resolution (CR 608.2b): a player who has left the game gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dbd6a94b-62ff-4a10-9d52-bdd90b26e425",
		Name:         "Time Warp",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    targetPlayerTakesExtraTurns(1),
	})
}

// targetPlayerTakesExtraTurns is the resolving body of "target player
// takes N extra turns after this one" (Time Warp, Time Stretch).
func targetPlayerTakesExtraTurns(n int) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetPlayer {
				return TakeExtraTurn{Player: t.ID, N: n}.Apply(ctx)
			}
		}
		return nil
	}
}
