package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silt Crawler — Creature — Beast {2}{G}, 3/3 (Prophecy, S82 OLD batch 674,
// #5617):
//
//	"When this creature enters, tap all lands you control."
//
// Troublesome Spirit's tap-every-land walk over the controller's own lands, on
// an enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ba1595a-6578-459d-96fe-fdfb84ac34d5",
		Name:         "Silt Crawler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Silt Crawler — tap all lands you control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B24TapAllLandsControlledBy(effects.NewContext(g, item), item.Controller)
				}),
		},
	})
}
