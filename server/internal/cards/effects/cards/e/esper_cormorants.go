package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Esper Cormorants — Artifact Creature — Bird {2}{W}{U}, 3/3 (Conflux,
// #5587):
//
//	"Flying"
//
// Flying is a printed keyword: it can be blocked only by creatures with
// flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aeb6cb00-e8bd-4610-aa39-29663d9c9160",
		Name:            "Esper Cormorants",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
