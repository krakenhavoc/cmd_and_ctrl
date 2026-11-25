package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rampart Smasher — Creature — Giant {R/G}{R/G}{R/G}{R/G}, 5/5 (Throne of
// Eldraine, S82 old sets):
//
//	"This creature can't be blocked by Knights or Walls."
//
// Juggernaut's block rule with two creature types (CR 509.1b): a creature that
// is a Knight or a Wall, as blockers are declared, can't block it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ece1778-fb19-40d5-8bcc-f23033e1d5c0",
		Name:         "Rampart Smasher",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.AnySubtype("Knight", "Wall"), "Knights or Walls"),
		},
	})
}
