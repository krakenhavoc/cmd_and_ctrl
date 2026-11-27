package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Behemoth — Creature — Elemental {3}{G}{G} 4/4 (Eighth Edition, S82 older sets batch 619, #5607):
//
//	"Tap an untapped creature you control: This creature gets +1/+1 until end
//	 of turn."
//
// Bramblesnap's pump: the cost is tapping any untapped creature you control
// (itself included), which is not a {T} cost, so summoning sickness does not
// stop it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4364bef8-f7ec-487c-9b64-b55c4afb864e",
		Name:         "Llanowar Behemoth",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped creature you control: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.TapAnUntapped("an untapped creature you control", effects.Creature()),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Llanowar Behemoth — +1/+1 until end of turn"),
		}},
	})
}
