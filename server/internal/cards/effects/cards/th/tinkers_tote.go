package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tinker's Tote — Artifact {2}{W} (The Lost Caverns of Ixalan, #4470):
//
//	"When this artifact enters, create two 1/1 colorless Gnome artifact
//	 creature tokens.
//	 {W}, Sacrifice this artifact: You gain 3 life."
//
// The enters trigger uses the stack and makes the table's 1/1 colorless
// Gnome artifact creature. The life gain sacrifices the Tote as its cost
// (CR 701.21a), so it can be activated in response to removal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd1e6225-31b7-4a70-81ab-681abe860328",
		Name:         "Tinker's Tote",
		Purpose:      game.Purpose{Tokens: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tinker's Tote — create two 1/1 colorless Gnome artifact creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Gnome artifact"), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, Sacrifice this artifact: You gain 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.GainLife{Amount: 3}),
		}},
	})
}
