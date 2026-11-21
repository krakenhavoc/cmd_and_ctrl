package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Watchful Blisterzoa — Creature — Phyrexian Jellyfish {4}{U}{U}, 4/4
// (Phyrexia: All Will Be One):
//
//	"Flying
//	 This creature enters with an oil counter on it.
//	 When this creature dies, draw cards equal to the number of oil counters
//	 on it."
//
// The count is read as the creature last existed on the battlefield
// (CR 603.10a), so one that spent its counter draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca130ba1-a4a8-4225-9799-75d56bac6fb3",
		Name:            "Watchful Blisterzoa",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OilCounter, 1, "Watchful Blisterzoa: enters with an oil counter on it")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Watchful Blisterzoa — draw cards equal to the number of oil counters on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					left, ok := ctx.TriggeringPermanent()
					if !ok {
						return nil
					}
					return effects.DrawCards{N: left.Counters[effects.OilCounter]}.Apply(ctx)
				}),
		},
	})
}
