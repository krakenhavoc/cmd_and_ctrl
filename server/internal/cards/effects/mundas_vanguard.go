package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Munda's Vanguard — Creature — Kor Knight Ally {4}{W}, 3/3 (Oath of the Gatewatch, #5545):
//
//	"Cohort — {T}, Tap an untapped Ally you control: Put a +1/+1
//	 counter on each creature you control."
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
		OracleID:     "1d7ee372-3a3d-4e23-a1ce-b8b8cdd4bdae",
		Name:         "Munda's Vanguard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Cohort — {T}, Tap an untapped Ally you control: Put a +1/+1 counter on each creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ogwCohortCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b11PutCountersOnEachCreatureYouControl(g, item, 1)
			},
		}},
	})
}
