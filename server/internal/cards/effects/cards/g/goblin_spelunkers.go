package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Spelunkers — Creature — Goblin Warrior {2}{R}, 2/2 (Ravnica: City of
// Guilds, S82 OLD b582):
//
//	"Mountainwalk (This creature can't be blocked as long as defending player
//	 controls a Mountain.)"
//
// Mountain Goat's mountainwalk: a printed keyword the block legality check
// reads (CR 702.14c), against the defending player's lands only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0534a514-2847-4dbc-8178-6a2886c75bf6",
		Name:            "Goblin Spelunkers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
