package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firescreamer — Creature — Kavu {3}{B}, 2/2 (Invasion, S82 OLD #5616):
//
//	"{R}: This creature gets +1/+0 until end of turn."
//
// Firebreathing's shape: a self boost until end of turn (CR 611.2c), as many
// times as the mana allows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b86c0c23-bb29-4af0-bbd4-50b3ae634375",
		Name:         "Firescreamer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Firescreamer — +1/+0 until end of turn"),
		}},
	})
}
