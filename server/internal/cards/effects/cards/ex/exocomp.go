package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exocomp — Artifact Creature — Construct {2}{U}, 3/1 (Star Trek
// Commander, #4974):
//
//	"Flying"
//
// The printed keyword, which the engine enforces (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb208157-67e5-45e2-bf08-7671f65cf520",
		Name:            "Exocomp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
