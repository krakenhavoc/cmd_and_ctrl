package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Landslide — {R} Sorcery (Urza's Destiny, S82 OLD batch 704, #5621):
//
//	"Sacrifice any number of Mountains. Landslide deals that much damage to
//	 target player or planeswalker."
//
// The Mountains are chosen and sacrificed as it resolves, and the damage is the
// number actually sacrificed (S82OldB704Landslide).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f15761cf-2ac5-4458-957c-15badb1dda50",
		Name:         "Landslide",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		OnResolve:    effects.S82OldB704Landslide,
	})
}
