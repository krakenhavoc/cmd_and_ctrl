package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Luminollusk — Creature {3}{G} (Lorwyn Eclipsed, #4458):
//
//	"Deathtouch
//	 Vivid — When this creature enters, you gain life equal to the number
//	 of colors among permanents you control."
//
// Vivid is an ability word with no rules meaning. The colors are counted
// as the ability resolves, the Luminollusk included (CR 105.2).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8bf2dce-7bcd-45cb-bf15-f53e24333a01",
		Name:            "Luminollusk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Luminollusk — you gain life equal to the number of colors among permanents you control (vivid)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Amount: len(effects.ColorsAmongPermanentsYouControl(g, item.Controller))}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
