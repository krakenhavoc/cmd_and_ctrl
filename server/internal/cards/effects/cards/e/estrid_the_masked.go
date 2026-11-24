package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Estrid, the Masked — Legendary Planeswalker — Estrid {1}{G}{W}{U}, loyalty 3
// (Secret Lair Drop, a reversible card whose two faces are the same card):
//
//	"+2: Untap each enchanted permanent you control.
//	 −1: Create a white Aura enchantment token named Mask attached to another
//	 target permanent. The token has enchant permanent and umbra armor.
//	 −7: Mill seven cards. Return all non-Aura enchantment cards from your
//	 graveyard to the battlefield, then do the same for Aura cards.
//	 Estrid, the Masked can be your commander."
//
// "Can be your commander" is read by the deck builder, not the engine.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7cf1fa2e-1572-4d9d-8e93-e49016b34458",
		Name:            "Estrid, the Masked",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+2: Untap each enchanted permanent you control.",
				Cost:   effects.LoyaltyCost(2),
				Effect: effects.OldB006UntapEachEnchantedPermanentYouControl,
			},
			{
				Label:   "−1: Create a white Aura enchantment token named Mask attached to another target permanent. The token has enchant permanent and umbra armor.",
				Cost:    effects.LoyaltyCost(-1),
				Targets: effects.Another(effects.TargetPermanent("another target permanent")),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					host, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					_, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.OldB006MaskToken()), 1,
						game.TokenEntryOptions{AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
					return err
				},
			},
			{
				Label:  "−7: Mill seven cards. Return all non-Aura enchantment cards from your graveyard to the battlefield, then do the same for Aura cards.",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.OldB006MillSevenReturnEnchantments,
			},
		},
	})
}
