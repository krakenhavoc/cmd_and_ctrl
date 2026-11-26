package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silent Submersible — Artifact — Vehicle {U}{U}, 2/3 (War of the Spark, #5520):
//
//	"Whenever this Vehicle deals combat damage to a player or planeswalker,
//	 draw a card.
//	 Crew 2"
//
// Mole Module's crew ability over a combat damage trigger on the player or
// planeswalker predicate Grateful Apparition uses. The draw is the
// Submersible's controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43257a85-2cbe-4931-9755-106808ab124a",
		Name:         "Silent Submersible",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Silent Submersible"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				"Silent Submersible — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
