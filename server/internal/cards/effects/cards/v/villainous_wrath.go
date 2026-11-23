package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "1c7c1867-20cb-4035-bb2a-42760edf0a55",
		Name:         "Villainous Wrath",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				if n := effects.CountControlled(ctx.Game, t.ID, effects.MatchCreature); n > 0 {
					if err := (effects.GainLife{Player: t.ID, Amount: -n}).Apply(ctx); err != nil {
						return err
					}
				}
				break
			}
			return effects.DestroyAllMatching{Match: effects.Creature()}.Apply(ctx)
		},
	})
}
