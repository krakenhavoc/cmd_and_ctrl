package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quiet Purity — Instant — Arcane {W} (Champions of Kamigawa, #5603):
//
//	"Destroy target enchantment."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f16dc14-3d3f-4ffa-90bc-9abc67db75cf",
		Name:         "Quiet Purity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
