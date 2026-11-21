package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shinestriker — Creature — Elemental {4}{U}{U}, 3/3 (Lorwyn Eclipsed,
// #4458):
//
//	"Flying
//	 Vivid — When this creature enters, draw cards equal to the number of
//	 colors among permanents you control."
//
// Vivid is an ability word with no rules meaning (CR 207.2c). The count is
// Explosive Prodigy's colorsAmongPermanentsYouControl, read as the trigger
// resolves, so the Shinestriker's own blue counts and a permanent that
// left in response does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b94ee87-a2f4-47f1-9fc7-4a167cad1b3b",
		Name:            "Shinestriker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shinestriker — draw cards equal to the number of colors among permanents you control (vivid)",
				func(g *game.Game, item *game.StackItem) error {
					n := len(effects.ColorsAmongPermanentsYouControl(g, item.Controller))
					return effects.DrawCards{N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
