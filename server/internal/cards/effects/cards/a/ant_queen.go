package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ant Queen — Creature — Insect {3}{G}{G}, 5/5:
//
//	"{1}{G}: Create a 1/1 green Insect creature token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50de88d4-0beb-45e5-9664-d4f0e280e652",
		Name:         "Ant Queen",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Create a 1/1 green Insect creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Insect"), N: 1}),
		}},
	})
}
