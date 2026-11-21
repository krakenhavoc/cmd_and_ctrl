package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bold Impaler — Creature — Vampire Knight {R}, 1/2 (Eldritch Moon, #5541):
//
//	"{2}{R}: This creature gets +2/+0 until end of turn."
//
// Hematite Golem's pump at {2}{R}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "374e6c1b-5a9a-4d3f-8bff-2f6bc0419391",
		Name:         "Bold Impaler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Bold Impaler — +2/+0 until end of turn"),
		}},
	})
}
