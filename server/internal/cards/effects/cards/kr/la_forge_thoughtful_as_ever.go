package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// La Forge, Thoughtful as Ever — {3}{U} Legendary Creature — Human
// Artificer 2/4 (Star Trek Commander, #4974):
//
//	"{T}: For each permanent you control that you've put a counter on
//	 this turn, create a 1/1 blue Faerie creature token with flying."
//
// The count is read as the ability resolves (CR 608.2h): permanents you
// control then that you put a counter of any kind on this turn, each
// object once however many counters, read off this turn's event log
// (socB25YouPutACounterOnThisObjectThisTurn, Fractal Tender's reader).
// Counters another player put, and counters on an earlier object of the
// same card (CR 400.7), don't count; a counter since removed still does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5e2ca35-053e-4155-97e7-73f75ae87d52",
		Name:         "La Forge, Thoughtful as Ever",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: For each permanent you control that you've put a counter on this turn, create a 1/1 blue Faerie creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := effects.TrekB46PermanentsYouControlYouPutACounterOnThisTurn(g, item.Controller)
				return effects.CreateToken{Template: effects.TokenCard(effects.TrekB46FaerieToken), N: n}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
