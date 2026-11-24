package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elemental Masterpiece — Sorcery {5}{U}{R} (Strixhaven: School of Mages,
// #5502):
//
//	"Create two 4/4 blue and red Elemental creature tokens.
//	 {U/R}{U/R}, Discard this card: Create a Treasure token."
//
// Two tokens from the printed 4/4 Elemental row, and a hand-zone activated
// ability like cycling and Channel: {U/R}{U/R} is two hybrid payments, the
// discard of this card is a cost (CR 602.2), and the ability creates one
// Treasure without casting anything.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9f5f87d-a485-4e69-9beb-49a51230bd53",
		Name:         "Elemental Masterpiece",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("4/4 blue and red Elemental"), N: 2}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U/R}{U/R}, Discard this card: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{U/R}{U/R}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
