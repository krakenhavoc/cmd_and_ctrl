package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metalspinner's Puzzleknot — Artifact {2} (Foundations Jumpstart):
//
//	"When this artifact enters, you draw a card and you lose 1 life.
//	 {2}{B}, Sacrifice this artifact: You draw a card and you lose 1 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c19d2f54-2c3a-4cbe-8328-061d4b51f574",
		Name:         "Metalspinner's Puzzleknot",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Metalspinner's Puzzleknot — you draw a card and you lose 1 life",
				Do(DrawCards{N: 1}, youLoseLife(1))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice this artifact: You draw a card and you lose 1 life.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{B}"), SacrificeThis()),
			Effect:  Do(DrawCards{N: 1}, youLoseLife(1)),
		}},
	})
}
