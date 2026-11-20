package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loreseeker's Stone — Artifact {6} (Commander 2014, #5553):
//
//	"{3}, {T}: Draw three cards. This ability costs {1} more to activate for
//	 each card in your hand."
//
// The increase is the ability's own cost modifier (Deepwood Denizen's
// shape, CR 602.2b and 601.2f), counted from the activating player's hand as
// the cost is determined, so the Stone is cheapest with an empty hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28a21e56-b97f-487d-9a89-61c248192cec",
		Name:         "Loreseeker's Stone",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Draw three cards. This ability costs {1} more to activate for each card in your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsMoreEach(loreseekersStoneCardsInHand,
					"This ability costs {1} more to activate for each card in your hand."),
			},
			Effect: Do(DrawCards{N: 3}),
		}},
	})
}

// loreseekersStoneCardsInHand is the number of cards in the activating
// player's hand.
func loreseekersStoneCardsInHand(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	if p := q.Game.PlayerByIDForEffect(q.Controller); p != nil {
		return p.Hand.Size()
	}
	return 0
}
