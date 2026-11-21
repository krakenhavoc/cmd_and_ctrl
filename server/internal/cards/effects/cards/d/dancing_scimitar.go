package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dancing Scimitar — Artifact Creature — Spirit {4}, 1/5 (Arabian
// Nights, #5641):
//
//	"Flying (This creature can't be blocked except by creatures with
//	 flying or reach.)"
//
// Flying is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "82b13601-d460-45c5-94a1-07e146d463a9",
		Name:            "Dancing Scimitar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
