package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// A.I.M. Bot — Artifact Creature — Robot Villain {2}{U}, 2/3 (Marvel Super
// Heroes Commander, #4455): "Flying". A printed keyword and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bd269d87-5f72-4559-bf2b-270e2ebe6154",
		Name:            "A.I.M. Bot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
