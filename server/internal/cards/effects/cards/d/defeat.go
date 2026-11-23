package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defeat — Sorcery {1}{B} (Dragons of Tarkir,
// #5551):
//
//	"Destroy target creature with power 2 or less."
//
// Swat's body without the cycling: power is read at announce and again
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21645448-ecbd-4036-b9ed-47cd903079e4",
		Name:         "Defeat",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
