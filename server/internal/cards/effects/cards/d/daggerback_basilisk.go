package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daggerback Basilisk — Creature — Basilisk {2}{G}, 2/2 (Rise of the
// Eldrazi, #5581):
//
//	"Deathtouch (Any amount of damage this deals to a creature is enough to
//	 destroy it.)"
//
// A keyword-only body; deathtouch is the engine's keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea4d30c7-7c39-46d5-9b71-a7ebc3e219a1",
		Name:            "Daggerback Basilisk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
