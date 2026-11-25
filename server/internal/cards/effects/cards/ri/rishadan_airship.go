package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rishadan Airship — Creature — Human Pirate {2}{U}, 3/1 (Jumpstart, #5510):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Battlefield Percher's block rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3c4661da-db2e-4834-9bba-d9a5592465d2",
		Name:            "Rishadan Airship",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      effects.NemBlocksOnlyFlyers(effects.OnSelf()),
	})
}
