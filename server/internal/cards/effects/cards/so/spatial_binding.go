package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spatial Binding — Enchantment {U}{B} (Mirage, S82 OLD batch 809, #5633):
//
//	"Pay 1 life: Until your next upkeep, target permanent can't phase out."
//
// The phasing seam of #6201: the record is pinned to the target for the
// activator's next upkeep. The life is a cost (CR 119.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52fc8217-5a14-40b5-a8fc-b6fc45af5f33",
		Name:         "Spatial Binding",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 1 life: Until your next upkeep, target permanent can't phase out.",
			Cost:    effects.PayLife(1),
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.TargetCantPhaseOutUntilYourNextUpkeep("Spatial Binding"),
		}},
	})
}
