package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terminate — Instant {B}{R} (EDHREC rank 224):
//
//	"Destroy target creature. It can't be regenerated."
//
// Unconditional creature removal at instant speed; Doom Blade
// without the colour clause.
//
// "It can't be regenerated" is ENFORCED as of #667 (CR 701.19c):
// the rider rides the destroy route onto the CR 614 event and the
// regeneration built-in declines. The shield is not spent
// (CR 701.19c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6257c2fd-005f-41e3-8a72-af76df1eb134",
		Name:         "Terminate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}
