package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Topan Ascetic — Creature — Human Monk {2}{G}, 2/2 (Shards of Alara, S82 OLD
// batch 511, #5588):
//
//	"Tap an untapped creature you control: This creature gets +1/+1 until end
//	 of turn."
//
// The tapped creature may be the Ascetic itself (TapAnUntapped has no
// "another"), and it need not be free of summoning sickness: this is not the
// {T} symbol (CR 302.6).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5868cf5-9fdc-442d-95cb-3bc5e2727981",
		Name:         "Topan Ascetic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped creature you control: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.TapAnUntapped("an untapped creature you control", effects.Creature()),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Topan Ascetic — +1/+1 until end of turn"),
		}},
	})
}
