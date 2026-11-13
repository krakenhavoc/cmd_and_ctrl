package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brass Secretary — Artifact Creature — Construct {3}, 2/1 (Urza's
// Destiny, #5621):
//
//	"{2}, Sacrifice this creature: Draw a card."
//
// Illvoi Galeblade's sacrifice-to-draw row on a colorless body. The
// sacrifice is part of the cost, so the card is drawn even though the
// Construct is gone when the ability resolves, and with no {T} in the
// cost it can be activated the turn the Construct arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d8c2d7b-7cce-4115-a3f8-18c23731544e",
		Name:         "Brass Secretary",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect:  ebDrawACard,
		}},
	})
}
