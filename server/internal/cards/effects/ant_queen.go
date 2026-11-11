package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ant Queen — Creature — Insect {3}{G}{G}, 5/5:
//
//	"{1}{G}: Create a 1/1 green Insect creature token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50de88d4-0beb-45e5-9664-d4f0e280e652",
		Name:         "Ant Queen",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Create a 1/1 green Insect creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{1}{G}"),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 green Insect"), N: 1}),
		}},
	})
}
