package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Lightning — Sorcery {3}{R}{R}{R} (Saviors of Kamigawa, #5601):
//
//	"Create three 3/1 red Elemental creature tokens with haste. Exile
//	 them at the beginning of the next end step."
//
// Chandra, Flamecaller's tokens and exile: the three tokens that actually
// landed are handed to the end-step exile through the creation's keyed tail
// (CR 603.7), so a token a replacement kept off the battlefield is simply not
// exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acd17fbf-645b-4c4d-b2c6-44127e3b743d",
		Name:         "Feral Lightning",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: ctx.Controller(),
				Source:     ctx.Source(),
				Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("3/1 red Elemental with haste")), Count: 3}},
			}, ctx.ItemTail(feralLightningExile))
		},
	})
}

// feralLightningExile is "exile them at the beginning of the next end step".
// An on-disk key.
var feralLightningExile = effects.SocB8ExileTokensAtNextEndStep("feral-lightning/exile-the-elementals",
	"Feral Lightning — exile the Elementals")
