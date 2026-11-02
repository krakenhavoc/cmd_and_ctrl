package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Villainous Wrath — Sorcery {3}{B}{B} (Marvel's Spider-Man, #4460):
//
//	"Target opponent loses life equal to the number of creatures they
//	 control. Then destroy all creatures."
//
// The count is read as the spell resolves, before the sweep. A target that is
// no longer legal fizzles the whole spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c7c1867-20cb-4035-bb2a-42760edf0a55",
		Name:         "Villainous Wrath",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				if n := countControlled(ctx.Game, t.ID, MatchCreature); n > 0 {
					if err := (GainLife{Player: t.ID, Amount: -n}).Apply(ctx); err != nil {
						return err
					}
				}
				break
			}
			return DestroyAllMatching{Match: Creature()}.Apply(ctx)
		},
	})
}
