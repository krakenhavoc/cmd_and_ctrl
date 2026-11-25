package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Freerunner — Creature — Goblin Warrior Ally {3}{R}, 3/2 (Oath of the
// Gatewatch, #5545):
//
//	"Surge {1}{R} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 Menace"
//
// Surge is the shared gated alternative cost (CR 702.117a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "372094f8-86dc-4c78-a5d0-444ffbf02630",
		Name:             "Goblin Freerunner",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{1}{R}")},
	})
}
