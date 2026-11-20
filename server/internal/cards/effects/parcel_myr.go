package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Parcel Myr — Artifact Creature — Clue Myr {1}{U}, 2/1 (Modern Horizons
// 2, #5501):
//
//	"{2}, Sacrifice this creature: Draw a card."
//
// Heart Warden's sacrifice-to-draw row. It has no {T}, so it can be
// activated the turn the Myr arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e660f66-0a94-4694-95e0-25296799b193",
		Name:         "Parcel Myr",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect:  ebDrawACard,
		}},
	})
}
