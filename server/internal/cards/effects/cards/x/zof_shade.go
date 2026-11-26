package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zof Shade — Creature — Shade {3}{B}, 2/2 (Magic 2015, #5556):
//
//	"{2}{B}: This creature gets +2/+2 until end of turn."
//
// Bellows Lizard's shape with a different cost and bonus: no {T}, so it can
// be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e4b8fe8-882c-4d0f-8fec-0256c62a0df0",
		Name:         "Zof Shade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    effects.ManaCost("{2}{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Zof Shade — +2/+2 until end of turn"),
		}},
	})
}
