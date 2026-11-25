package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sram's Expertise — {2}{W}{W} Sorcery (Aether Revolt, #5536):
//
//	"Create three 1/1 colorless Servo artifact creature tokens.
//	 You may cast a spell with mana value 3 or less from your hand without
//	 paying its mana cost."
//
// The tokens come first, as printed; the free cast is offered once they
// are on the battlefield (Baral's Expertise's offer).
//
// No simplification.
var sramsExpertiseFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/srams-expertise-cast", "Sram's Expertise")

func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cd17347-5558-46f9-b8bb-ae5deeb1f229",
		Name:         "Sram's Expertise",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 3}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				effects.OfferFreeHandCast(ctx, ctx.Controller(), sramsExpertiseFreeCast,
					"a spell with mana value 3 or less", effects.SpellWithManaValueAtMost(3))
				return nil
			}),
		),
	})
}
