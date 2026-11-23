package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steelshaper Apprentice — {2}{W}{W} Creature — Human Soldier:
//
//	"{W}, {T}, Return this creature to its owner's hand: Search your
//	 library for an Equipment card, reveal that card, put it into your
//	 hand, then shuffle."
//
// #2028: the return is the cost (ReturnThis), paid with the mana and
// the tap at announce (CR 602.2b). The search is Steelshaper's Gift's
// (b06TutorToHand).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf325b3a-0b28-4660-8b51-4334b59a9034",
		Name:         "Steelshaper Apprentice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}, Return this creature to its owner's hand: Search your library for an Equipment card, reveal that card, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost(), effects.ReturnThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B06TutorToHand("Steelshaper Apprentice — an Equipment card, revealed, to hand", effects.B09IsEquipmentCard)(item, effects.NewContext(g, item))
			},
		}},
	})
}
