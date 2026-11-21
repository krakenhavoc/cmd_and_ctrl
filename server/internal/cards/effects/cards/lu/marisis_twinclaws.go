package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marisi's Twinclaws — Creature — Cat Warrior {2}{R/W}{G}, 2/4 (Alara
// Reborn, #5586):
//
//	"Double strike"
//
// The printed keyword the engine enforces (CR 702.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1fa3839a-127e-4740-bd2e-97b18fdddedc",
		Name:            "Marisi's Twinclaws",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
