package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scaled Wurm — Creature — Wurm {7}{G}, 7/6 (Ice Age, #5636):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "009a8e38-74d0-4b4a-b8ca-62c9f3183531",
		Name:         "Scaled Wurm",
		Completeness: effects.CompletenessFull,
	})
}
