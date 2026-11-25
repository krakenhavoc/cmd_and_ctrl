package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruinous Gremlin — Creature — Gremlin {R}, 1/1 (Kaladesh, #5539):
//
//	"{2}{R}, Sacrifice this creature: Destroy target artifact."
//
// Yavimaya Steelcrusher's destroy with the sacrifice of the creature itself
// as a cost (CR 602.2b); with the target gone at resolution it does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b88d67c0-ef12-453c-8e58-6642650dffcc",
		Name:         "Ruinous Gremlin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this creature: Destroy target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
