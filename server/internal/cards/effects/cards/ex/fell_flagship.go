package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fell Flagship — Artifact — Vehicle {3}, 3/3 (Ixalan, S82 older sets):
//
//	"Pirates you control get +1/+0.
//	 Whenever this Vehicle deals combat damage to a player, that player
//	 discards a card.
//	 Crew 3"
//
// The anthem is a tribal static (Admiral Beckett Brass's, without "other"; the
// Vehicle is no Pirate unless it is made one), the discard is Banshee of the
// Dread Choir's, and the crew is the plain crew ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5afa88a-8bde-47fe-a890-3759f807360f",
		Name:         "Fell Flagship",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Pirate"}, YoursOnly: true}, 1, 0)},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Fell Flagship — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DiscardOne(g, item, item.Trigger.Event.Target)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Fell Flagship"),
		}},
	})
}
