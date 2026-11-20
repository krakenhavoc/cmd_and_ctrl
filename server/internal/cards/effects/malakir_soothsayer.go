package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malakir Soothsayer — Creature — Vampire Shaman Ally {4}{B}, 4/4 (Oath of the Gatewatch, #5545):
//
//	"Cohort — {T}, Tap an untapped Ally you control: You draw a card
//	 and you lose 1 life."
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
		OracleID:     "6ebfce76-82a2-4168-bd64-99cf16882dee",
		Name:         "Malakir Soothsayer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Cohort — {T}, Tap an untapped Ally you control: You draw a card and you lose 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    ogwCohortCost(),
			Effect:  Do(DrawCards{N: 1}, youLoseLife(1)),
		}},
	})
}
