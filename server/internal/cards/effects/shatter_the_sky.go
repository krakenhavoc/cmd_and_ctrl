package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shatter the Sky — Sorcery {2}{W}{W}:
//
//	"Each player who controls a creature with power 4 or greater draws a
//	 card. Then destroy all creatures."
//
// The draw comes first and is decided then: every player at the table who
// controls a creature with layered power 4 or greater draws one, in turn
// order from the caster, and only afterwards (ResolveSteps waits out a
// prompt a draw raises) are the creatures destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ffc7509-c7ce-44c4-a8cf-90ed161b5d10",
		Name:         "Shatter the Sky",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				drawers := MatchingBattlefield(ctx, And(Creature(), PowerGE(4)))
				for _, p := range tablePlayers(ctx) {
					for _, c := range drawers {
						if c.Controller == p {
							if err := (DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
								return err
							}
							break
						}
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				return DestroyAllMatching{Match: Creature()}.Apply(ctx)
			}),
		),
	})
}
