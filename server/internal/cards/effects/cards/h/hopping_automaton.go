package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hopping Automaton — Artifact Creature — Construct {3} 2/2 (Urza's Saga,
// #5624):
//
//	"{0}: This creature gets -1/-1 and gains flying until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7848e040-e1b4-4f81-922e-b1eca8fe4398",
		Name:         "Hopping Automaton",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: This creature gets -1/-1 and gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{0}"),
			Effect:  effects.ThisCreatureUntilEOT("Hopping Automaton — -1/-1 and flying until end of turn", -1, -1, "flying"),
		}},
	})
}
