package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rend Flesh — Instant — Arcane {2}{B} (Champions of Kamigawa, #5603):
//
//	"Destroy target non-Spirit creature."
//
// "Non-Spirit" is the creature type on the permanent as it is now, so a
// creature that becomes a Spirit in response is no longer a legal target
// and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "079d4698-dc11-4d29-91b6-725d78dfb75d",
		Name:         "Rend Flesh",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target non-Spirit creature", effects.Creature(), effects.Not(effects.HasSubtype("Spirit"))),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
