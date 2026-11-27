package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gulf Squid — Creature — Squid Beast {3}{U}, 2/2 (Prophecy, S82 OLD batch 674,
// #5617):
//
//	"When this creature enters, tap all lands target player controls."
//
// Mana Short's tap-every-land walk (B24TapAllLandsControlledBy) on an enters
// trigger that targets a player, any player, the controller included. A player
// who left before resolution is an illegal target (CR 608.2b) and nothing
// happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98a60a97-3a43-4544-8ea3-40d0bad79caf",
		Name:         "Gulf Squid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Gulf Squid — tap all lands target player controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return effects.B24TapAllLandsControlledBy(ctx, target)
				}),
				effects.TargetPlayer("target player")),
		},
	})
}
