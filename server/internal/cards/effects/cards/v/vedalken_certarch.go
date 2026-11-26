package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Certarch — Creature — Vedalken Wizard {U}, 1/1 (Scars of Mirrodin,
// S82 OLD batch 469, #5577):
//
//	"Metalcraft — {T}: Tap target artifact, creature, or land. Activate only
//	 if you control three or more artifacts."
//
// Ring of Gix's tap body on a {T} cost, gated as the ability is activated
// (CR 602.5b); the {T} is a creature's, so the Certarch waits out summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "337533fb-2783-493c-856f-c41a83cca443",
		Name:         "Vedalken Certarch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "Metalcraft — {T}: Tap target artifact, creature, or land. Activate only if you control three or more artifacts.",
			Cost:      effects.TapCost(),
			Condition: effects.OldB469MetalcraftActivation(),
			Targets:   effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
			Effect:    effects.TapChosenPermanent,
		}},
	})
}
