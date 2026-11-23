package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moriok Replica — Artifact Creature — Warrior {3}, 2/2 (Scars of Mirrodin, #5577):
//
//	"{1}{B}, Sacrifice this creature: You draw two cards and you lose 2 life."
//
// Infernal Idol's draw-two-lose-two, sacrificing the creature as the
// cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6273cd72-5330-4712-ae73-b058023f0435",
		Name:         "Moriok Replica",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice this creature: You draw two cards and you lose 2 life.",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Effect:  effects.Do(effects.DrawCards{N: 2}, effects.YouLoseLife(2)),
		}},
	})
}
