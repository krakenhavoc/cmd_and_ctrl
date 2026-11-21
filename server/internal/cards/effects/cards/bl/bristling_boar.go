package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bristling Boar — Creature — Boar {3}{G}, 4/3 (Core Set 2019, #5527):
//
//	"This creature can't be blocked by more than one creature."
//
// One blocker at most (CR 509.1b), the MaxBlockers block rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4aed3d2-04f5-49a4-a6be-d3cf097903f8",
		Name:         "Bristling Boar",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.MaxBlockers(effects.OnSelf(), 1),
		},
	})
}
