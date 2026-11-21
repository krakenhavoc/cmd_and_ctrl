package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bounding Wolf — Creature — Wolf {2}{G}, 3/2 (Innistrad: Midnight Hunt, S82):
//
//	"Flash
//	 Reach"
//
// Keywords only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "257f7c29-8857-429e-a421-9262982804e0",
		Name:            "Bounding Wolf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "reach"},
	})
}
