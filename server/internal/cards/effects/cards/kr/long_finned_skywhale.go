package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Long-Finned Skywhale — Creature — Whale {2}{U}{U}, 4/3 (Kaladesh, #5539):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Flying is printed; the block restriction is Battlefield Percher's
// (nemBlocksOnlyFlyers), judged as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a8287efd-23ca-498b-90f1-c76099919fa4",
		Name:            "Long-Finned Skywhale",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      effects.NemBlocksOnlyFlyers(effects.OnSelf()),
	})
}
