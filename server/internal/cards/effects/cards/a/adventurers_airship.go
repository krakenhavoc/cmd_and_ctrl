package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "e3016295-00f7-4daf-91b0-57653733b103",
		Name:            "Adventurer's Airship",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Adventurer's Airship — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 1) }),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Adventurer's Airship"),
		}},
	})
}
