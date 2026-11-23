package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormfront Pegasus — Creature — Pegasus {1}{W}, 2/1 (Magic 2010, #5585):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b35c1b30-c1dc-489f-bd79-ad6566fb9389",
		Name:            "Stormfront Pegasus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
