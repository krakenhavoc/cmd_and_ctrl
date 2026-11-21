package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinister Starfish — Creature — Starfish {1}{B}, 0/3:
//
//	"{T}: Surveil 1."
//
// A tap ability whose whole effect is one surveil, as D'Vana Tendi's
// {3} ability is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58321467-41d6-4419-bcb2-cd3e60f677b3",
		Name:         "Sinister Starfish",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.Surveil{N: 1}),
		}},
	})
}
