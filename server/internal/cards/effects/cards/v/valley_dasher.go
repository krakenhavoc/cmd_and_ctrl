package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "aad16b50-a866-4871-843f-ec8813ad5dfd",
		Name:            "Valley Dasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
