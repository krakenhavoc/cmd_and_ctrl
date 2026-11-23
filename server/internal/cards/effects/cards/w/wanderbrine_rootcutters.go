package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wanderbrine Rootcutters — Creature — Merfolk Rogue {2}{U/B}{U/B}, 3/3
// (Shadowmoor, #5591):
//
//	"This creature can't be blocked by green creatures."
//
// A block rule, Shield Mare's: a blocker of that colour, as it is now, is
// refused (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e2e96db-e4df-41de-b5a7-712c4a8fe7be",
		Name:         "Wanderbrine Rootcutters",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.OfColor("G")), "green creatures"),
		},
	})
}
