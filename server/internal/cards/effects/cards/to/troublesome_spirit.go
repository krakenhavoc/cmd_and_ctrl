package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Troublesome Spirit — Creature — Spirit {2}{U}{U}, 3/4 (Prophecy, S82 OLD batch
// 673, #5617):
//
//	"Flying
//	 At the beginning of your end step, tap all lands you control."
//
// Flying is a printed keyword. The trigger is Mana Short's tap-every-land walk
// over the controller's own lands, at the beginning of their end step only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "601d9bc3-ca75-43de-beea-ab3f0a49f543",
		Name:            "Troublesome Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.ByYou, "Troublesome Spirit — tap all lands you control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B24TapAllLandsControlledBy(effects.NewContext(g, item), item.Controller)
				}),
		},
	})
}
