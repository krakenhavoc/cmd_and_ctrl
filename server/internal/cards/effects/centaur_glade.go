package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Centaur Glade — Enchantment {3}{G}{G} (Onslaught):
//
//	"{2}{G}{G}: Create a 3/3 green Centaur creature token."
//
// Ant Queen's shape on an enchantment: no tap, so it can be activated
// any number of times at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "369c3b6c-6715-4b6a-89d2-e7542400ba8f",
		Name:         "Centaur Glade",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}{G}: Create a 3/3 green Centaur creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{2}{G}{G}"),
			Effect:  Do(CreateToken{Template: TokenCard("3/3 green Centaur"), N: 1}),
		}},
	})
}
