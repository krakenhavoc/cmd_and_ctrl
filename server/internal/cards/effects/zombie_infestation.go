package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zombie Infestation — Enchantment {1}{B} (Jumpstart):
//
//	"Discard two cards: Create a 2/2 black Zombie creature token."
//
// The discard is the cost, paid in full as the ability is activated: a
// hand of fewer than two cannot activate it. No tap and no mana, so it
// may be activated any number of times at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdce1af0-3643-4e77-88f9-320206a191d4",
		Name:         "Zombie Infestation",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard two cards: Create a 2/2 black Zombie creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Cost:    DiscardN(2, "two cards"),
			Effect:  Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}),
		}},
	})
}
