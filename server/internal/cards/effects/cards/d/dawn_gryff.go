package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawn Gryff — Creature — Hippogriff {2}{W}, 2/2 (Eldritch Moon, #5541):
//
//	"Flying"
//
// Snapping Drake's keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1c34b2e4-e600-4688-86cb-8a1c11fdc770",
		Name:            "Dawn Gryff",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
