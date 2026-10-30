package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Big Wheel — Artifact — Vehicle {2}{R}, 4/4 (Marvel Super Heroes
// Commander, #4455):
//
//	"Trample
//	 When this Vehicle enters, you may discard a card. If you do, draw a
//	 card.
//	 Crew 2"
//
// The loot is Improvised Act's and Battlefield Scavenger's shared body
// (b39MayDiscardThenDraw); the crew ability is the ordinary printed one.
func init() {
	const label = "Big Wheel — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:        "b0941bc2-f482-4821-a96b-720b08c7334a",
		Name:            "Big Wheel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, label, drawAsManyAsDiscarded)(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Big Wheel"),
		}},
	})
}
