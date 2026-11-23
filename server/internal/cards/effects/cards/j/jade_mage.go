package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jade Mage — Creature — Human Shaman {1}{G}, 2/1 (Magic 2012, #5573):
//
//	"{2}{G}: Create a 1/1 green Saproling creature token."
//
// Selesnya Guildmage's Saproling row at {2}{G}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6daf3c59-7636-4d44-872b-67722b4868cb",
		Name:         "Jade Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{2}{G}"),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}),
		}},
	})
}
