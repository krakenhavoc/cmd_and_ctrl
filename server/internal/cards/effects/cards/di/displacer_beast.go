package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Displacer Beast — Creature — Cat Beast {2}{U}, 3/2:
//
//	"When this creature enters, venture into the dungeon.
//	 Displacement — {3}{U}: Return this creature to its owner's hand."
//
// Bouncing it and casting it again ventures again, which is the card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f24c0b7-1d39-48b6-969f-ee97fdf99305",
		Name:         "Displacer Beast",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.WhenThisEntersVenture("Displacer Beast")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Displacement — {3}{U}: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
