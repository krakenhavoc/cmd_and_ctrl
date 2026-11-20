package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moriok Replica — Artifact Creature — Warrior {3}, 2/2 (Scars of Mirrodin, #5577):
//
//	"{1}{B}, Sacrifice this creature: You draw two cards and you lose 2 life."
//
// Infernal Idol's draw-two-lose-two, sacrificing the creature as the
// cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6273cd72-5330-4712-ae73-b058023f0435",
		Name:         "Moriok Replica",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice this creature: You draw two cards and you lose 2 life.",
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Effect:  Do(DrawCards{N: 2}, youLoseLife(2)),
		}},
	})
}
