package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soramaro, First to Dream — Legendary Creature — Spirit {4}{U}{U}, */*
// (Saviors of Kamigawa, #5601):
//
//	"Flying
//	 Soramaro's power and toughness are each equal to the number of cards in
//	 your hand.
//	 {4}, Return a land you control to its owner's hand: Draw a card."
//
// Maro's characteristic-defining ability over the controller's hand,
// counted live, and Moonbow Illusionist's cost: the land comes back as the
// cost is paid (CR 602.2b), tapped or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b11f96a0-5f70-45d6-90cc-2edbead3672e",
		Name:            "Soramaro, First to Dream",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.S82OldB021CardsInHandPT("Soramaro's power and toughness are each equal to the number of cards in your hand."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, Return a land you control to its owner's hand: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.ReturnAPermanentToHand("a land you control", effects.Land())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
