package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pearl Dragon — Creature — Dragon {4}{W}{W}, 4/4 (Mirage, #5633):
//
//	"Flying
//	 {1}{W}: This creature gets +0/+1 until end of turn."
//
// Flying is the printed keyword. The pump is the shared
// thisGetsUntilEndOfTurn body at +0/+1, any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f4e0e2e0-8d64-40a3-92da-3951a696b5f3",
		Name:            "Pearl Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: This creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(0, 1, "Pearl Dragon — +0/+1 until end of turn"),
		}},
	})
}
