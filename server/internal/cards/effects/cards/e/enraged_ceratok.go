package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enraged Ceratok — Creature — Rhino {2}{G}{G}, 4/4 (Ravnica: Clue Edition):
//
//	"This creature can't be blocked by creatures with power 2 or less."
//
// Chomping Kavu's block rule on the creature itself, judged on each
// would-be blocker's current power as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1f3f377-c7b1-474c-aed6-78b061d138c1",
		Name:         "Enraged Ceratok",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
