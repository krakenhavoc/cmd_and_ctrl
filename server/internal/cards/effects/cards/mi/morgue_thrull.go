package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morgue Thrull — Creature — Thrull {2}{B}, 2/2 (Stronghold, #5627):
//
//	"Sacrifice this creature: Mill three cards."
//
// The sacrifice is the cost (CR 701.21a), so it can be activated at any
// time it has priority, summoning sick or not, and in response to its own
// removal. Its controller mills three (CR 701.17a, the shared MillCards
// primitive), or as many as the library has (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc759e09-34b2-4004-8474-a6500dd1280c",
		Name:         "Morgue Thrull",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Mill three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.SacrificeThis(),
			Effect:  effects.Do(effects.MillCards{N: 3}),
		}},
	})
}
