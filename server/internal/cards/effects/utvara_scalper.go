package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Utvara Scalper — Creature — Goblin Scout {1}{R}, 1/2 (Ravnica: Clue
// Edition, S82):
//
//	"Flying
//	 This creature attacks each combat if able."
//
// Barricade Breaker's attack requirement on a flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "58e754ad-88b4-47c5-b105-8b50e54143c8",
		Name:            "Utvara Scalper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
	})
}
