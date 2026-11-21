package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Air Elemental — Creature — Elemental {3}{U}{U}, 4/4
// (Limited Edition Alpha, #5642): "Flying". A printed keyword and
// nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7744bae4-a8b7-44a5-9b4c-0048ad4cc448",
		Name:            "Air Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
