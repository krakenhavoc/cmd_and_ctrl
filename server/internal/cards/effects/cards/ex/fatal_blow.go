package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fatal Blow — Instant {B} (Weatherlight, #5629):
//
//	"Destroy target creature that was dealt damage this turn. It can't be
//	 regenerated."
//
// Qutrub Forayer's target clause. The destruction ignores a
// regeneration shield (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16774611-c004-4383-b4e1-4a43a9b1f4aa",
		Name:         "Fatal Blow",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature that was dealt damage this turn", effects.DealtDamageThisTurn()),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}
