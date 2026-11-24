package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karn, Scion of Urza — Legendary Planeswalker — Karn {4}, loyalty 5:
//
//	"+1: Reveal the top two cards of your library. An opponent chooses one of
//	 them. Put that card into your hand and exile the other with a silver
//	 counter on it.
//	 −1: Put a card you own with a silver counter on it from exile into your
//	 hand.
//	 −2: Create a 0/0 colorless Construct artifact creature token with "This
//	 token gets +1/+1 for each artifact you control.""
//
// The opponent who chooses is named by the controller (RevealPick, the
// Gifts Ungiven shape). The silver counter is on the exiled card itself, so
// the −1 reads any card the controller owns in exile with one, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "060209fa-4b7d-4e08-bf8b-8759bba5641b",
		Name:            "Karn, Scion of Urza",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Reveal the top two cards of your library. An opponent chooses one of them. Put that card into your hand and exile the other with a silver counter on it.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.S82OldB005CKarnScionPlusOne,
			},
			{
				Label:  "−1: Put a card you own with a silver counter on it from exile into your hand.",
				Cost:   effects.LoyaltyCost(-1),
				Effect: effects.S82OldB005CKarnScionMinusOne,
			},
			{
				Label:   "−2: Create a 0/0 colorless Construct artifact creature token with \"This token gets +1/+1 for each artifact you control.\"",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Template: effects.UrzaConstructToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
