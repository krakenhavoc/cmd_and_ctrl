package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Spider — "Reach."
//
// Canonical reach ground-blocker. 2/4 for 3G. The keyword feeds
// Game.CanBlockLocked: Giant Spider can block flying attackers (Serra Angel,
// Vampire Nighthawk).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e740ce2f-2134-473c-afa1-1b6d2d1e38ef",
		Name:            "Giant Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
