package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spawnbinder Mage — Creature — Human Wizard Ally {3}{W}, 2/4 (Oath of the Gatewatch, #5545):
//
//	"Cohort — {T}, Tap an untapped Ally you control: Tap target
//	 creature."
//
// Cohort is an ability word (CR 207.2c): the cost is this creature's {T}
// plus tapping an untapped Ally you control (the tap-others cost,
// game/tap_others_cost.go), which is not the {T} symbol, so an Ally that
// arrived this turn may pay it. The creature can't pay both halves (its
// own {T} has tapped it), so the clause excludes it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d49c50b-3e2d-4e74-9b8a-c2f9a495e470",
		Name:         "Spawnbinder Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Cohort — {T}, Tap an untapped Ally you control: Tap target creature.",
			Cost:    effects.OgwCohortCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
