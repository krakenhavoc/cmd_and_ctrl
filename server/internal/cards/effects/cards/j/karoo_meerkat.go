package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karoo Meerkat — Creature — Mongoose {1}{G}, 2/1 (Mirage, S82 OLD b806):
//
//	"Protection from blue"
//
// One protection quality (CR 702.16), carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "871549e6-24c0-4fbd-8b71-ba8639b0ee47",
		Name:            "Karoo Meerkat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.ProtectionFromColor("U")},
	})
}
