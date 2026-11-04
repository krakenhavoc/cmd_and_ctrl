package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Subway Train — Artifact — Vehicle {2}, 3/1 (Marvel's Spider-Man, #4460):
//
//	"When this Vehicle enters, you may pay {G}. If you do, search your
//	 library for a basic land card, reveal it, put it into your hand, then
//	 shuffle.
//	 Crew 2"
//
// The {G} is asked as the trigger resolves (MayPay, Ant-Man's shape) and the
// search runs only when it was really paid. Crew 2 is Magmatic Galleon's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d4472ad-3fbf-4a5d-94de-dfee04fda9c3",
		Name:         "Subway Train",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Subway Train — you may pay {G} to search for a basic land card", subwayTrainOffer),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Subway Train"),
		}},
	})
}

// subwayTrainOffer asks for the {G}; the search lives in the paid branch.
func subwayTrainOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{G}",
		Question: "Subway Train — pay {G} to search your library for a basic land card?",
		Then:     subwayTrainPaid,
	}.Apply(NewContext(g, item))
}

// subwayTrainPaid is the paid half, keyed: reveal a basic land card into
// the hand, then shuffle.
var subwayTrainPaid = PayUnlessBranches("pay-unless/subway-train-basic-land",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return SearchLibrary{
			Player:    r.Chooser,
			Predicate: IsBasicLand,
			Dest:      game.ZoneHand,
			Limit:     1,
			Reveal:    true,
			Shuffle:   true,
			Reason:    "Subway Train — a basic land card, revealed and put into your hand",
		}.Apply(ctx)
	}, nil)
