package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vessel of Ephemera — Enchantment {1}{W} (Shadows over Innistrad, S82 OLD):
//
//	"{2}{W}, Sacrifice this enchantment: Create two 1/1 white Spirit creature
//	 tokens with flying."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfddc488-ab38-4e6d-8dfc-22bf1a8e03fa",
		Name:         "Vessel of Ephemera",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, Sacrifice this enchantment: Create two 1/1 white Spirit creature tokens with flying.",
			Purpose: game.Purpose{Tokens: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.B28WhiteSpiritFlyingToken(), N: 2}),
		}},
	})
}
