package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind Drill Assailant — Creature — Rat Warlock {2}{U/B}{U/B}, 2/5
// (Bloomburrow, #4467):
//
//	"Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +3/+0.
//	 {2}{U/B}: Surveil 1."
//
// Threshold is an ability word with no rules of its own (CR 207.2c): the
// layer 7c self-modifier (CR 613.4c) the Reality Fracture threshold cards
// use, read live on every recompute. The surveil is an ordinary activated
// ability that may be activated any time you could cast an instant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b10ab39-3050-475c-af8b-f293794d498c",
		Name:         "Mind Drill Assailant",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.FraThresholdSelfPT(3, 0)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U/B}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{U/B}"),
			Effect:  effects.Do(effects.Surveil{N: 1}),
		}},
	})
}
