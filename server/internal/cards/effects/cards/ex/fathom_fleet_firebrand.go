package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fathom Fleet Firebrand — Creature — Human Pirate {1}{R}, 2/2:
//
//	"{1}{R}: This creature gets +1/+0 until end of turn."
//
// Firebreathing through the shared thisGetsUntilEndOfTurn body: no {T}, so it
// can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a297f9a2-3461-4138-8a3e-75027cbdbe12",
		Name:         "Fathom Fleet Firebrand",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Fathom Fleet Firebrand — +1/+0 until end of turn"),
		}},
	})
}
