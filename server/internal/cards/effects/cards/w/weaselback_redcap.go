package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weaselback Redcap — Creature — Goblin Knight {R}, 1/1 (Foundations
// Jumpstart, S82, #5929):
//
//	"{1}{R}: This creature gets +2/+0 until end of turn."
//
// Freejam Regent's firebreathing activation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b14bd895-1d4e-4bd5-9ac1-fed56afb6c90",
		Name:         "Weaselback Redcap",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Weaselback Redcap — +2/+0"),
		}},
	})
}
