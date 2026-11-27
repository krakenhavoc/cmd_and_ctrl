package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Huang Zhong, Shu General — Legendary Creature — Human Soldier {2}{W}{W}, 2/3
// (Portal Three Kingdoms, #5622):
//
//	"Huang Zhong can't be blocked by more than one creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d6c4290-d46d-4b98-805c-3f537462c4c8",
		Name:         "Huang Zhong, Shu General",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.MaxBlockers(effects.OnSelf(), 1),
		},
	})
}
