package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ajani's Sunstriker — Creature — Cat Cleric {W}{W}, 2/2 (Magic 2013, #5568):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain that
//	 much life.)"
//
// A keyword body: lifelink (CR 702.15) rides the printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "31009c45-afba-45f4-a6ca-9dfdb9990e72",
		Name:            "Ajani's Sunstriker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
