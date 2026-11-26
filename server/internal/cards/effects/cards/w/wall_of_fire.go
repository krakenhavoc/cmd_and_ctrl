package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Fire — Creature — Wall {1}{R}{R}, 0/5 (Magic 2015, #5556):
//
//	"Defender
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Bellows Lizard's firebreathing at {R}, on a defender.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f38c8b47-e8e0-4d2f-b1da-d8d986805a48",
		Name:            "Wall of Fire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Wall of Fire — +1/+0 until end of turn"),
		}},
	})
}
