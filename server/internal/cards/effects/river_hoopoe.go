package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "058bbc0a-9c09-4c98-8735-83255a71bbca",
		Name:            "River Hoopoe",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}{U}: You gain 2 life and draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, LifeGain: 2},
			Cost:    ManaCost("{3}{G}{U}"),
			Effect:  Do(GainLife{Amount: 2}, DrawCards{N: 1}),
		}},
	})
}
