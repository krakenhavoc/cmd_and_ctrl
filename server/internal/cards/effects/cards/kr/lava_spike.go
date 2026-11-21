package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lava Spike — Sorcery — Arcane {R} (Champions of Kamigawa, #5603):
//
//	"Lava Spike deals 3 damage to target player or planeswalker."
//
// Skullcrack's target clause without its riders. An Arcane spell, so it
// can carry a splice (CR 702.47).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2837888c-bfa8-4955-9334-0605ad409f7e",
		Name:         "Lava Spike",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		OnResolve:    effects.DamageToFirstTargetStep(3),
	})
}
