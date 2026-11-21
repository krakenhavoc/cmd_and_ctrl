package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Claws of Gix — Artifact {0} (Urza's Saga):
//
//	"{1}, Sacrifice a permanent: You gain 1 life."
//
// A sacrifice outlet for any permanent you control, the Claws included.
// It uses the stack (it is not a mana ability), and the sacrifice is a
// cost, so it happens even if the ability is countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4d384d7-f294-4b2d-9971-a4689c150255",
		Name:         "Claws of Gix",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a permanent: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAPermanent()),
			Effect:  effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
}
