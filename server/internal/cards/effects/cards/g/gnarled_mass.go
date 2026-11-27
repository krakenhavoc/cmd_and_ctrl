package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnarled Mass — Creature — Spirit {1}{G}{G}, 3/3 (Betrayers of Kamigawa, S82 OLD b595):
// no rules text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18202185-f9d4-446a-b34e-fae9800240f9",
		Name:         "Gnarled Mass",
		Completeness: effects.CompletenessFull,
	})
}
