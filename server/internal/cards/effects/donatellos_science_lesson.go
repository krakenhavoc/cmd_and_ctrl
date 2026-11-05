package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Donatello's Science Lesson — Instant {2}{U} (TMNT Commander, #5011):
//
//	"Tap up to two target creatures. Up to two target players each draw
//	 a card."
//
// Two target clauses read by slot: zero to two creatures, then zero to
// two players. Each target still legal as the spell resolves is acted
// on (CR 608.2b); the creatures are tapped first, then each player
// draws, in the order they were chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd61a01c-dfe1-4c1e-8063-7f90fcd0c452",
		Name:         "Donatello's Science Lesson",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("up to two target creatures").WithCount(0, 2),
			TargetPlayer("up to two target players").WithCount(0, 2),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, t := range ctx.ClauseTargets(1) {
				if t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (DrawCards{Player: t.ID, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
