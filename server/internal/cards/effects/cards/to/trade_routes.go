package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trade Routes — Enchantment {1}{U} (Eighth Edition, S82 older sets batch 618, #5607):
//
//	"{1}: Return target land you control to its owner's hand.
//	 {1}, Discard a land card: Draw a card."
//
// Vedalken Mastermind's bounce of a permanent you control, narrowed to a land,
// and Molten Vortex's discard-a-land cost with a draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "120fa67e-e5e0-4c23-9a78-d6171d357aee",
		Name:         "Trade Routes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}: Return target land you control to its owner's hand.",
				Cost:    effects.ManaCost("{1}"),
				Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
				Effect:  effects.BounceTheTargetToHand,
			},
			{
				Label:   "{1}, Discard a land card: Draw a card.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue}, // ADR 0142: a draw answers nothing on the stack
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.DiscardCardsMatching(1, "a land card", game.Card.IsLand)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
