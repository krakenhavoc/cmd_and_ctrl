package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Staff of the Ages — Artifact {3} (Ice Age, S82 OLD batch 847, #5636):
//
//	"Creatures with landwalk abilities can be blocked as though they didn't have
//	 those abilities."
//
// The block rule of #6195 (BlockableAsThoughNoLandwalkAtAll): every landwalk
// keyword, the snow family included, is waived for every attacking creature at
// the table while the Staff is on the battlefield. The creatures keep the
// keyword, and every other evasion applies as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86af21ba-58ad-45e8-a316-5411c870e415",
		Name:         "Staff of the Ages",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.BlockableAsThoughNoLandwalkAtAll()},
	})
}
