package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drana's Chosen — Creature — Vampire Shaman Ally {3}{B}, 2/2 (Oath of the Gatewatch, #5545):
//
//	"Cohort — {T}, Tap an untapped Ally you control: Create a tapped
//	 2/2 black Zombie creature token."
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
		OracleID:     "185b3b9f-f798-4ca4-af99-b17a24fcc1c6",
		Name:         "Drana's Chosen",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Cohort — {T}, Tap an untapped Ally you control: Create a tapped 2/2 black Zombie creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.OgwCohortCost(),
			Effect:  effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.BlackZombieToken()).EntersTapped(), N: 1}),
		}},
	})
}
