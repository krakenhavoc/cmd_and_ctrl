package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valley Dasher — Creature — Human Berserker {1}{R}, 2/2 (Khans of
// Tarkir, #5554):
//
//	"Haste
//	 This creature attacks each combat if able."
//
// Haste is printed data (CR 702.10b); the attack requirement is
// Juggernaut's AttacksEachCombat static (CR 508.1d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aad16b50-a866-4871-843f-ec8813ad5dfd",
		Name:            "Valley Dasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
	})
}
