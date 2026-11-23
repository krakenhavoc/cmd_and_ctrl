package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slinking Skirge — Creature — Phyrexian Imp {3}{B}, 2/1 (Urza's
// Destiny, #5621):
//
//	"Flying
//	 {2}, Sacrifice this creature: Draw a card."
//
// Illvoi Galeblade without flash: a flier that cashes itself in for a
// card. The sacrifice is paid with the {2} at activation, so the draw
// happens even though the Imp has already left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "82875793-b264-4ceb-8525-ef3b6d086072",
		Name:            "Slinking Skirge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Effect:  effects.EbDrawACard,
		}},
	})
}
