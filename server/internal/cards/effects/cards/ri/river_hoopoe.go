package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// River Hoopoe — Creature — Bird {G}{U}, 1/3 (Hour of Devastation, #5534):
//
//	"Flying
//	 {3}{G}{U}: You gain 2 life and draw a card."
//
// Flying is a printed keyword. The ability is two plain steps in printed
// order, the life gain and then the draw, and can be activated any number
// of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "058bbc0a-9c09-4c98-8735-83255a71bbca",
		Name:            "River Hoopoe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}{U}: You gain 2 life and draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, LifeGain: 2},
			Cost:    effects.ManaCost("{3}{G}{U}"),
			Effect:  effects.Do(effects.GainLife{Amount: 2}, effects.DrawCards{N: 1}),
		}},
	})
}
