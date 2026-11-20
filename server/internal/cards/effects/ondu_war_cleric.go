package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ondu War Cleric — Creature — Human Cleric Ally {1}{W}, 2/2 (Oath of the Gatewatch, #5545):
//
//	"Cohort — {T}, Tap an untapped Ally you control: You gain 2 life."
//
// Cohort is an ability word (CR 207.2c): the cost is this creature's {T}
// plus tapping an untapped Ally you control (the tap-others cost,
// game/tap_others_cost.go), which is not the {T} symbol, so an Ally that
// arrived this turn may pay it. The creature can't pay both halves (its
// own {T} has tapped it), so the clause excludes it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d011153e-d12a-4eb1-b105-64729e37be12",
		Name:         "Ondu War Cleric",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Cohort — {T}, Tap an untapped Ally you control: You gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ogwCohortCost(),
			Effect:  Do(GainLife{Amount: 2}),
		}},
	})
}
