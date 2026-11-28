package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Vortex — Instant {1}{U}{B} (Ice Age, S82 OLD, #5636):
//
//	"Destroy target creature unless its controller pays life equal to its toughness. A creature
//	 destroyed this way can't be regenerated."
//
// The creature's controller (not the caster) is asked as the spell resolves, with the toughness read
// then through the layers; a controller who can't pay that much life (CR 119.4) loses the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea38a211-a7b0-49fe-b5dc-7eca7610cccb",
		Name:         "Essence Vortex",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.S82OldB842EssenceVortexResolve,
	})
}
