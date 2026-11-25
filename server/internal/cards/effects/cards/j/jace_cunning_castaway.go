package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace, Cunning Castaway — Legendary Planeswalker — Jace {1}{U}{U},
// loyalty 3 (Ixalan, #5532):
//
//	"+1: Whenever one or more creatures you control deal combat damage to a
//	     player this turn, draw a card, then discard a card.
//	 −2: Create a 2/2 blue Illusion creature token with "When this creature
//	     becomes the target of a spell, sacrifice it."
//	 −5: Create two tokens that are copies of Jace, except they're not
//	     legendary."
//
// The +1 creates a repeating delayed trigger (CR 603.7b) that fires once for
// each player damaged in a combat damage step however many of your creatures
// connect (CR 603.2c), and keeps going if Jace leaves. The −2's Illusion
// sacrifices itself the moment a spell targets it. The −5 copies Jace
// (or his last-known self) twice, each copy entering with his printed
// loyalty and no longer legendary (CR 707.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d2341a16-ffde-42fa-9746-0d8888e812d4",
		Name:            "Jace, Cunning Castaway",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Whenever one or more creatures you control deal combat damage to a player this turn, draw a card, then discard a card.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect:  effects.OldB228JacePlusOne,
			},
			{
				Label:   "−2: Create a 2/2 blue Illusion creature token with \"When this creature becomes the target of a spell, sacrifice it.\"",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.OldB228IllusionToken(), N: 1}),
			},
			{
				Label:   "−5: Create two tokens that are copies of Jace, except they're not legendary.",
				Cost:    effects.LoyaltyCost(-5),
				Purpose: game.Purpose{Tokens: 2},
				Effect:  effects.OldB228JaceCopies,
			},
		},
	})
}
