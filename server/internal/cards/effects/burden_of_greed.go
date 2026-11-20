package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burden of Greed — Instant {3}{B} (Darksteel, #5605):
//
//	"Target player loses 1 life for each tapped artifact they control."
//
// Guiltfeeder's life loss with the count read as the spell resolves (CR
// 608.2h): the tapped artifacts the target controls then. With none it
// does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a80c13db-949d-4784-8177-04aaa22b96b0",
		Name:         "Burden of Greed",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				n := len(permanentsControlledByMatching(ctx.Game, t.ID, And(Artifact(), Not(Untapped()))))
				if n <= 0 {
					return nil
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -n)
			}
			return nil
		},
	})
}
