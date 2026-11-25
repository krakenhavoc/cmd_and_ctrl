package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renegade Freighter — Artifact — Vehicle {3}, 4/3 (Kaladesh, #5539):
//
//	"Whenever this Vehicle attacks, it gets +1/+1 and gains trample until end
//	 of turn.
//	 Crew 2"
//
// Crew 2 is the ordinary CR 702.122 crew ability (Ribskiff's). The attack
// trigger is the source's own until-end-of-turn pump and trample grant, which
// refuse a source that has left and come back as a new object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e73c8d2-5cad-4201-b9ef-32be82797e3b",
		Name:         "Renegade Freighter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Renegade Freighter — it gets +1/+1 and gains trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					const label = "Renegade Freighter — +1/+1 and trample until end of turn"
					if err := effects.ThisGetsUntilEndOfTurn(1, 1, label)(g, item); err != nil {
						return err
					}
					return effects.ThisGainsKeywordUntilEndOfTurn("trample", label)(g, item)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Renegade Freighter"),
		}},
	})
}
