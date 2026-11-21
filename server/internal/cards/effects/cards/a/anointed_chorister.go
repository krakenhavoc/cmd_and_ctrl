package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anointed Chorister — Creature — Human Cleric {W}, 1/1 (Core Set 2021, #5511):
//
//	"Lifelink
//	 {4}{W}: This creature gets +3/+3 until end of turn."
//
// Lifelink rides PrintedKeywords; the pump is thisGetsRow, a layer 7c
// boost on the source that ends at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "61859443-e298-41c7-a179-6e6554de91c9",
		Name:            "Anointed Chorister",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{4}{W}", 3, 3, "Anointed Chorister")},
	})
}
