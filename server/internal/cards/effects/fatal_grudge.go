package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:       "37114872-9804-4ee8-a443-21ea6eaddb0c",
		Name:           "Fatal Grudge",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a nonland permanent", Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			info, ok := ctx.SacrificedPermanent()
			if !ok {
				return DrawCards{N: 1}.Apply(ctx)
			}
			return EachPlayerSacrifices{
				ExceptController: true,
				Label:            "a permanent that shares a card type with the sacrificed permanent",
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return sharesACardTypeWith(info, c)
				},
				Then: func(ctx *Context, _ game.PromptedSacrifices) error {
					return DrawCards{N: 1}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
