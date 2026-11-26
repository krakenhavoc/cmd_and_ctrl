package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akoum Flameseeker — Creature — Human Shaman Ally {2}{R}, 3/2 (Oath of the
// Gatewatch, #5545):
//
//	"Cohort — {T}, Tap an untapped Ally you control: Discard a card. If you do,
//	 draw a card."
//
// Cohort is an ability word (CR 207.2c): the cost is the Flameseeker's {T}
// plus tapping another untapped Ally (OgwCohortCost, Munda's Vanguard's cost),
// which is not the {T} symbol. The loot is Gristle Glutton's: with an empty hand
// nothing is discarded and nothing is drawn ("if you do").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efae637f-3232-46f2-9839-f3386e2f447d",
		Name:         "Akoum Flameseeker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Cohort — {T}, Tap an untapped Ally you control: Discard a card. If you do, draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Cost:    effects.OgwCohortCost(),
			Effect:  effects.DiscardACardIfYouDoDraw,
		}},
	})
}
