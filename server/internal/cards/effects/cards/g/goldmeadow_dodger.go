package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goldmeadow Dodger — Creature — Kithkin Rogue {W}, 1/1 (Lorwyn, S82 OLD,
// #5593):
//
//	"This creature can't be blocked by creatures with power 4 or greater."
//
// Brushfire Elemental's block rule with the other bound: read against each
// would-be blocker's power as blocks are declared (CR 509.1b), after the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e120f341-4445-4b96-9777-2290471832bf",
		Name:         "Goldmeadow Dodger",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerGE(4), "creatures with power 4 or greater"),
		},
	})
}
