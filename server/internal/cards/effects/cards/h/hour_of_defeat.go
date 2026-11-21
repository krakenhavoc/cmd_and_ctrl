package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hour of Defeat — Instant {3}{B} (Marvel Super Heroes, #4455):
//
//	"Destroy target creature. Surveil 1."
//
// Surveil happens whether or not the creature was destroyed (an
// indestructible target survives and the player still surveils). If the
// one target is illegal on resolution the spell does not resolve at all
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2774fcb7-8988-4a18-b5ec-7534dd8b2e05",
		Name:         "Hour of Defeat",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTargetThenSurveil(1),
	})
}
