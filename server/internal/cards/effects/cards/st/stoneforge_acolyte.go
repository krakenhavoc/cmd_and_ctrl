package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stoneforge Acolyte — Creature — Kor Artificer Ally {W}, 1/2 (Oath of the
// Gatewatch, #5545):
//
//	"Cohort — {T}, Tap an untapped Ally you control: Look at the top four cards
//	 of your library. You may reveal an Equipment card from among them and put
//	 it into your hand. Put the rest on the bottom of your library in any
//	 order."
//
// Cohort is an ability word (CR 207.2c); the cost is OgwCohortCost. The look is
// Dáin's Company's, with the rest ordered by the player (TakeRevealRestBottomAny)
// rather than at random.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3c40f93-7e10-4139-b9e0-b6dfff8b8038",
		Name:         "Stoneforge Acolyte",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Cohort — {T}, Tap an untapped Ally you control: Look at the top four cards of your library. You may reveal an Equipment card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    effects.OgwCohortCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := item.Controller
				return effects.TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
					Match:    effects.Subtype("Equipment"),
					Max:      1,
					Optional: true,
					Label:    "Stoneforge Acolyte — you may reveal an Equipment card and put it into your hand",
					Key:      effects.TakeRevealRestBottomAny,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
