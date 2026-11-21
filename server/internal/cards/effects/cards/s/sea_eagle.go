package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea Eagle — Creature — Bird {1}{U}, 1/1
// (Starter 1999, #5620):
//
//	"Flying"
//
// Flying is printed data: it can be blocked only by creatures with
// flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "acb57162-7093-4a3c-9818-d3b61ce757c6",
		Name:            "Sea Eagle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
