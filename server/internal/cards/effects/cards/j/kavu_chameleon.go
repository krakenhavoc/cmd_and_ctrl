package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kavu Chameleon (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Kavu {3}{G}{G}, 4/4:
//	"This spell can't be countered.
//	{G}: This creature becomes the color of your choice until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "82c583ab-8158-4074-a3ff-b7a05ecc8ab1",
		Name:            "Kavu Chameleon",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: This creature becomes the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{G}"),
			Effect:  effects.S82OldB646ThisBecomesChosenColor("Kavu Chameleon"),
		}},
	})
}
