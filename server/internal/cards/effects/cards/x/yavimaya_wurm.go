package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yavimaya Wurm — Creature — Wurm {4}{G}{G}, 6/4 (Magic 2011, #5579):
//
//	"Trample"
//
// A printed keyword and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "93514874-41e8-49e7-8935-d06e852dd1e3",
		Name:            "Yavimaya Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
