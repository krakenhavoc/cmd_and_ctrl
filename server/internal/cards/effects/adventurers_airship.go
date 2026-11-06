package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adventurer's Airship — Artifact — Vehicle {3}, 3/2 (Final Fantasy,
// #4462):
//
//	"Flying
//	 Whenever this Vehicle attacks, draw a card, then discard a card.
//	 Crew 2"
//
// The attack trigger is the shared loot body (b16DrawThenDiscard): the
// discard is chosen after the draw. Crew 2 is the ordinary CR 702.122
// crew ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e3016295-00f7-4daf-91b0-57653733b103",
		Name:            "Adventurer's Airship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Adventurer's Airship — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 1) }),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Adventurer's Airship"),
		}},
	})
}
