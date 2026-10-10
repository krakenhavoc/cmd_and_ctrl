package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sam, Loyal Attendant — Legendary Creature — Halfling Peasant
// {1}{G}{W}, 2/4:
//
//	"Partner with Frodo, Adventurous Hobbit (When this creature enters,
//	 target player may put Frodo into their hand from their library,
//	 then shuffle.)
//	 At the beginning of combat on your turn, create a Food token.
//	 Activated abilities of Foods you control cost {1} less to
//	 activate."
//
// Partner with is CR 702.124j's two abilities: the commander pairing
// with Frodo (internal/deck) and the entry search (PartnerWith, #2142).
//
// The discount is Boom Scholar's activation cost pass (#1184) with the
// source read as the Food: it reaches every Food you control, token or
// card, and a Food's {2} becomes {1}. It reduces generic mana only
// (CR 601.2f), and two Sams take a Food's {2} to {0}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2837be10-23ad-49d2-943e-b1e00edbbbd9",
		Name:         "Sam, Loyal Attendant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Sam, Loyal Attendant", "Frodo, Adventurous Hobbit"),
			AtBeginningOfYourCombat("Sam, Loyal Attendant — create a Food token",
				Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		CostModifiers: []game.CostModifier{
			ActivationCostsLess(1,
				"Activated abilities of Foods you control cost {1} less to activate.",
				ofAFoodYouControl()),
		},
	})
}

// ofAFoodYouControl — "activated abilities of Foods you control": the
// ability's source is a Food under the modifier's controller.
func ofAFoodYouControl() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.Controller == q.Source.Controller && q.Card.HasSubtype("Food")
	}
}
