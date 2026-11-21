package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Welder Automaton — Artifact Creature — Construct {2}, 2/1 (Aether Revolt,
// #5536):
//
//	"{3}{R}: This creature deals 1 damage to each opponent."
//
// Fire Nation Archers' DamageEachOpponent without the token: the
// Automaton is the source, one damage event per opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52eae443-fbe5-479f-a5a4-b5e4401d62d3",
		Name:         "Welder Automaton",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: This creature deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{R}"),
			Effect:  effects.Do(effects.DamageEachOpponent{N: 1}),
		}},
	})
}
