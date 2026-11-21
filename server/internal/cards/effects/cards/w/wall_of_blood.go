package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Blood — Creature — Wall {2}{B}, 0/2 (Mirrodin, #5606):
//
//	"Defender (This creature can't attack.)
//	 Pay 1 life: This creature gets +1/+1 until end of turn."
//
// Printed defender (CR 702.3b), and Carrion Howler's life-paid pump at
// +1/+1: a layer 7c boost on the source (CR 613.4c) that ends at
// cleanup. Life is paid as the cost (CR 119.4), so a player at 0 or
// less life can't pay it, and it can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b26c3563-d289-4ca3-8736-9d89e191a840",
		Name:            "Wall of Blood",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 1 life: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.PayLife(1),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Wall of Blood — +1/+1 until end of turn"),
		}},
	})
}
