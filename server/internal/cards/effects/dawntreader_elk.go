package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawntreader Elk — Creature — Elk {1}{G}, 2/2 (Dark Ascension, #5571):
//
//	"{G}, Sacrifice this creature: Search your library for a basic
//	 land card, put it onto the battlefield tapped, then shuffle."
//
// Evolving Wilds' fetch (fetchBasicTapped) on a creature: the Elk is
// sacrificed as the cost (CR 602.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de6d0248-79f6-48df-bdca-8ce207b2ec08",
		Name:         "Dawntreader Elk",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, Sacrifice this creature: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
			Cost:    Plus(ManaCost("{G}"), SacrificeThis()),
			Effect:  fetchBasicTapped,
		}},
	})
}
