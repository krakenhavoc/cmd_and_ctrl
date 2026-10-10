package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bf325b3a-0b28-4660-8b51-4334b59a9034",
		Name:         "Steelshaper Apprentice",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}, Return this creature to its owner's hand: Search your library for an Equipment card, reveal that card, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    Plus(ManaCost("{W}"), TapCost(), ReturnThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b06TutorToHand("Steelshaper Apprentice — an Equipment card, revealed, to hand", b09IsEquipmentCard)(item, NewContext(g, item))
			},
		}},
	})
}
