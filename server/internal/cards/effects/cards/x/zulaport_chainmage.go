package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zulaport Chainmage — Creature — Human Shaman Ally {3}{B}, 4/2 (Oath of the Gatewatch, #5545):
//
//	"Cohort — {T}, Tap an untapped Ally you control: Target opponent
//	 loses 2 life."
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
		OracleID:     "017f4471-db5a-4ad0-97b3-b1f12b52578a",
		Name:         "Zulaport Chainmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Cohort — {T}, Tap an untapped Ally you control: Target opponent loses 2 life.",
			Cost:    effects.OgwCohortCost(),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 2}),
		}},
	})
}
