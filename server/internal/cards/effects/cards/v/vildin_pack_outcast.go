package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vildin-Pack Outcast // Dronepack Kindred — {4}{R} Creature — Werewolf
// Horror 4/4 // Creature — Eldrazi Werewolf 5/7 (Eldritch Moon, S82 OLD,
// #5541):
//
//	Front: "Trample
//	        {R}: This creature gets +1/-1 until end of turn.
//	        {5}{R}{R}: Transform this creature."
//	Back:  "Trample
//	        {1}: This creature gets +1/+0 until end of turn."
//
// Flowstone Wall's firebreathing-with-a-cost and Shrill Howler's in-place
// transform at instant speed.
//
// No simplification.
func init() {
	const oracle = "6e6f9269-c72b-4449-9c2d-a720b9bd6f06"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Vildin-Pack Outcast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: This creature gets +1/-1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, -1, "Vildin-Pack Outcast — +1/-1 until end of turn"),
			},
			{
				Label:   "{5}{R}{R}: Transform this creature.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{5}{R}{R}"),
				Effect:  effects.Do(effects.TransformThis{}),
			},
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Dronepack Kindred",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{1}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Dronepack Kindred — +1/+0 until end of turn"),
		}},
	})
}
