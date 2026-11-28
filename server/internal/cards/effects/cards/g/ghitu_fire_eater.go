package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghitu Fire-Eater — Creature — Human Nomad {2}{R} 2/2:
//
//	"{T}, Sacrifice this creature: It deals damage equal to its power to any
//	 target."
//
// The power is read as the creature last existed on the battlefield, after the
// sacrifice paid the cost (CR 608.2h).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e792338-4a21-4a06-ad94-39931118e06c",
		Name:         "Ghitu Fire-Eater",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: It deals damage equal to its power to any target.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsDamageEqualToItsPowerToTarget,
		}},
	})
}
