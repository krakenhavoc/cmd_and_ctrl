package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Vanguard — Creature — Dryad {2}{G}, 1/1 (Invasion, S82 OLD #5616):
//
//	"{T}: This creature gets +0/+4 until end of turn."
//
// A tap ability on the creature itself, so summoning sickness bars it (CR
// 302.6); the boost is until end of turn (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9123086e-ae67-46c5-aed5-6366e2fe5a8c",
		Name:         "Llanowar Vanguard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature gets +0/+4 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 4}},
			Cost:    effects.TapCost(),
			Effect:  effects.ThisGetsUntilEndOfTurn(0, 4, "Llanowar Vanguard — +0/+4 until end of turn"),
		}},
	})
}
