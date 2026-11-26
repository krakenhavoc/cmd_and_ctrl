package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloud Elemental — Creature — Elemental {2}{U}, 2/3 (Magic 2011, #5579):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Flying is a printed keyword; the block restriction is the Percher's block
// rule (CR 509.1b): it may not be declared as a blocker of a creature
// without flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b751cc7-3a19-4207-9e64-83e18d279588",
		Name:            "Cloud Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      effects.NemBlocksOnlyFlyers(effects.OnSelf()),
	})
}
