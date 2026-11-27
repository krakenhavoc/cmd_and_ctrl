package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloud Sprite — Creature — Faerie {U}, 1/1 (Mercadian Masques, S82 OLD b683,
// #5619):
//
//	"Flying (This creature can't be blocked except by creatures with flying
//	 or reach.)
//	 This creature can block only creatures with flying."
//
// Flying is a printed keyword; the block restriction is Battlefield Percher's
// (NemBlocksOnlyFlyers on itself), a restriction judged as blockers are
// declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "283433cb-9466-4b40-8af9-30934654d43b",
		Name:            "Cloud Sprite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      effects.NemBlocksOnlyFlyers(effects.OnSelf()),
	})
}
