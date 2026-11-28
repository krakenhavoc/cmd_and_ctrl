package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pygmy Pyrosaur — Creature — Lizard {1}{R}, 1/1 (Urza's Legacy, S82 OLD b715, #5623):
//
//	"This creature can't block.
//	{R}: This creature gets +1/+0 until end of turn."
//
// Aesthir Glider's "can't block" restriction printed on itself, and
// Bellows Lizard's firebreathing at {R}: no {T}, so it can be activated any
// number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "667c39bc-92d6-4fee-bced-95f0e1f21455",
		Name:         "Pygmy Pyrosaur",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Pygmy Pyrosaur — +1/+0 until end of turn"),
		}},
	})
}
