package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rusted Relic — Artifact {4} (Scars of Mirrodin, S82 OLD batch 469, #5577):
//
//	"Metalcraft — This artifact is a 5/5 Golem artifact creature as long as
//	 you control three or more artifacts."
//
// A layer 4 type change (Creature, Golem) and a layer 7b base 5/5, both on
// the metalcraft gate and both turning on and off as the count crosses three
// (CR 611.3a). The Relic is an artifact itself, so it counts towards its own
// three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57b22bc3-23e8-4632-b10a-495fac5c77c7",
		Name:         "Rusted Relic",
		Completeness: effects.CompletenessFull,
		Static:       effects.OldB469SelfBecomesGolem(5, 5),
	})
}
