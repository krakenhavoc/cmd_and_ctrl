package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleet-Footed Monk — Creature — Human Monk {1}{W}, 1/1 (Portal, #5630):
//
//	"This creature can't be blocked by creatures with power 2 or greater."
//
// An evasion restriction, judged on each would-be blocker's power as
// blockers are declared (CR 509.1b), as Steel Leaf Champion's is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6a26261-1915-4699-a2e9-cca43b39a3eb",
		Name:         "Fleet-Footed Monk",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerGE(2), "creatures with power 2 or greater"),
		},
	})
}
