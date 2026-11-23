package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fatal Grudge — Sorcery {B}{R}:
//
//	"As an additional cost to cast this spell, sacrifice a nonland
//	 permanent.
//	 Each opponent chooses a permanent they control that shares a card
//	 type with the sacrificed permanent and sacrifices it.
//	 Draw a card."
//
// The card types are the sacrificed permanent's as it last existed on
// the battlefield (CR 608.2h), read off the payment record (ADR 0113
// §1); a copy uses the original's (the 2022-04-29 ruling, CR 707.10).
// Card types only (CR 205.2a): an artifact creature shares with any
// artifact and any creature. Each opponent chooses their own; one with
// nothing that shares a type sacrifices nothing. Then the caster draws.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "37114872-9804-4ee8-a443-21ea6eaddb0c",
		Name:           "Fatal Grudge",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a nonland permanent", effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			info, ok := ctx.SacrificedPermanent()
			if !ok {
				return effects.DrawCards{N: 1}.Apply(ctx)
			}
			return effects.EachPlayerSacrifices{
				ExceptController: true,
				Label:            "a permanent that shares a card type with the sacrificed permanent",
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return effects.SharesACardTypeWith(info, c)
				},
				Keyed: fatalGrudgeEdicted,
			}.Apply(ctx)
		},
	})
}

// fatalGrudgeEdicted is the card drawn once the edict has landed, keyed
// (#4290).
var fatalGrudgeEdicted = effects.RegisterEdictThen("prompt-run/fatal-grudge-draw", func(ctx *effects.Context, _ game.PromptedSacrifices) error {
	return effects.DrawCards{N: 1}.Apply(ctx)
})
