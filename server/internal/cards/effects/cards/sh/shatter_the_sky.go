package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2ffc7509-c7ce-44c4-a8cf-90ed161b5d10",
		Name:         "Shatter the Sky",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				drawers := effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.PowerGE(4)))
				for _, p := range effects.TablePlayers(ctx) {
					for _, c := range drawers {
						if c.Controller == p {
							if err := (effects.DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
								return err
							}
							break
						}
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DestroyAllMatching{Match: effects.Creature()}.Apply(ctx)
			}),
		),
	})
}
