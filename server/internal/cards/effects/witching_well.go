package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Witching Well — Artifact {U} (Foundations Jumpstart):
//
//	"When this artifact enters, scry 2.
//	 {3}{U}, Sacrifice this artifact: Draw two cards."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "baadfc17-42fd-4771-aabe-027ef7b3bf69",
		Name:         "Witching Well",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Witching Well — scry 2", Do(Scry{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Draws: 2, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}{U}"), SacrificeThis()),
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
