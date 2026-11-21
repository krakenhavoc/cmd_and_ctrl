package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mesa Falcon — Creature — Bird {1}{W}, 1/1 (Homelands, #5635):
//
//	"Flying
//	 {1}{W}: This creature gets +0/+1 until end of turn."
//
// Flying is an engine keyword carried as printed data (CR 702.9b). The
// boost is a layer 7c change (CR 613.4c) that ends at cleanup
// (CR 514.2); with no {T} in the cost it can be activated any number of
// times, including the turn the Falcon arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bc4c5f33-0f3d-4815-9974-fb12ac725d08",
		Name:            "Mesa Falcon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: This creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(0, 1, "Mesa Falcon — +0/+1 until end of turn"),
		}},
	})
}
