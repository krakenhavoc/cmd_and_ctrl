package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wit's End — Sorcery {5}{B}{B} (Dissension, #5598):
//
//	"Target player discards their hand."
//
// The whole hand goes at once (discardWholeHand); there is nothing to
// choose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97f11aed-7a5c-4a14-8c2c-5c230c1f05b4",
		Name:         "Wit's End",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				_, err := discardWholeHand(ctx.Game, t.ID)
				return err
			}
			return nil
		},
	})
}
