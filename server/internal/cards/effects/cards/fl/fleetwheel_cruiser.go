package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fleetwheel Cruiser — Artifact — Vehicle {4}, 5/3 (Kaladesh, #5539):
//
//	"Trample, haste
//	 When this Vehicle enters, it becomes an artifact creature until end of
//	 turn.
//	 Crew 2"
//
// The entry trigger is a Vehicle's own animation (Bespoke Battlewagon's
// BecomeCreatureUntilEOT on the source), the same until-end-of-turn change a
// crew ability makes, so with its printed haste it can attack the turn it
// arrives. Crew 2 is the ordinary crew ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b6b9c5a-d2ba-48b9-8efc-a3003c610130",
		Name:            "Fleetwheel Cruiser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fleetwheel Cruiser — it becomes an artifact creature until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if effects.SourceIsNewObject(g, item) {
						return nil
					}
					return effects.BecomeCreatureUntilEOT{Label: "Fleetwheel Cruiser — becomes an artifact creature"}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Fleetwheel Cruiser"),
		}},
	})
}
