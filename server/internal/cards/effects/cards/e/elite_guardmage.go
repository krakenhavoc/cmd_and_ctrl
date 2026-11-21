package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elite Guardmage — Creature — Human Wizard {2}{W}{U}, 2/3 (War of the Spark, #5520):
//
//	"Flying
//	 When this creature enters, you gain 3 life and draw a card."
//
// The life and the draw are two steps of one enters trigger, in printed
// order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "92dfeeb2-1117-422b-87ea-08a589f1134b",
		Name:            "Elite Guardmage",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{LifeGain: 3, Draws: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Elite Guardmage — you gain 3 life and draw a card",
				effects.Do(effects.GainLife{Amount: 3}, effects.DrawCards{N: 1})),
		},
	})
}
