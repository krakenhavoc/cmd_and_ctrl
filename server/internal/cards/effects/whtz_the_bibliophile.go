package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whtz, the Bibliophile — Legendary Creature — Homunculus {1}{W}{U}, 1/3
// (Mystery Booster Commander, #5948):
//
//	"Rulebreaker — A deck with this commander has no maximum deck size.
//	 {3}, {T}: You draw a card and gain 1 life. This ability costs {3} less to
//	 activate if you had 200 or more cards in your starting deck."
//
// The ability is a plain tap-and-pay draw. Its discount is an ability cost
// modifier that reads Player.StartingDeckSize, recorded when the deck was
// seated (the library plus the commanders, never a companion or a sideboard
// card), so it does not change as cards are drawn. At 200 the ability costs
// nothing but the {T}. The Rulebreaker sentence is deck construction and is
// read from the oracle text by the deck validator (internal/deck/rulebreaker.go):
// such a deck still holds at least 100 cards.
//
// No simplification.
func init() {
	const label = "{3}, {T}: You draw a card and gain 1 life. This ability costs {3} less to activate if you had 200 or more cards in your starting deck."
	Register(Spec{
		OracleID:     "73ba7ed0-0f08-4ba4-9f51-42c48f1b62ea",
		Name:         "Whtz, the Bibliophile",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(gapB05BigStartingDeckDiscount,
					"This ability costs {3} less to activate if you had 200 or more cards in your starting deck."),
			},
			Effect: Do(DrawCards{N: 1}, GainLife{Amount: 1}),
		}},
	})
}

// gapB05BigStartingDeckDiscount is "{3} less if you had 200 or more cards in
// your starting deck": 3 for the activating player with such a deck, else 0.
func gapB05BigStartingDeckDiscount(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	p := q.Game.PlayerByIDForEffect(q.Controller)
	if p == nil || p.StartingDeckSize < 200 {
		return 0
	}
	return 3
}
