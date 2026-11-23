package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yotian Medic — Creature — Human Cleric Soldier {2}{W}, 1/4 (The Brothers' War, #5482):
//
//	"Lifelink"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d218b9ad-3a6f-4bda-90c6-3f869c6c4e6d",
		Name:            "Yotian Medic",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
