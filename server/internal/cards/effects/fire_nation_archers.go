package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Nation Archers — Creature — Human Archer {3}{R}, 3/4 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"Reach
//	 {5}: This creature deals 2 damage to each opponent. Create a 2/2 red
//	 Soldier creature token."
//
// The Soldier is the Eternal set's plain 2/2 red Soldier, not the
// firebending one of the main set.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2e735849-1930-4f43-86b4-da660ade15bc",
		Name:            "Fire Nation Archers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{5}: This creature deals 2 damage to each opponent. Create a 2/2 red Soldier creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{5}"),
			Effect:  Do(DamageEachOpponent{N: 2}, CreateToken{Template: TokenCard("2/2 red Soldier"), N: 1}),
		}},
	})
}
