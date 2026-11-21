package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shifting Sliver — Creature — Sliver {3}{U}, 2/2 (Time Spiral remastered):
//
//	"Slivers can't be blocked except by Slivers."
//
// Prowler's Helm's "Walls can't be blocked except by Walls" with Slivers on
// both sides: every Sliver at the table, this one and an opponent's
// included, is refused any blocker that is not a Sliver (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a262f53-2eea-45d0-8189-9222d66f1b52",
		Name:         "Shifting Sliver",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnMatching(effects.HasSubtype("Sliver")), effects.HasSubtype("Sliver"), "Slivers"),
		},
	})
}
