package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Flamecaller — Legendary Planeswalker — Chandra {4}{R}{R}, loyalty 4:
//
//	"+1: Create two 3/1 red Elemental creature tokens with haste. Exile them at
//	 the beginning of the next end step.
//	 0: Discard all the cards in your hand, then draw that many cards plus one.
//	 −X: Chandra deals X damage to each creature."
//
// +1 exiles the two tokens that landed (the keyed tail of the creation).
// 0 counts the hand as it is discarded, then draws one more than that, so an
// empty hand draws one. −X is one simultaneous damage event to every creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c2a9131-f3d7-4f71-8bcc-3c169574b2e3",
		Name:            "Chandra, Flamecaller",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		XMatters:        true,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create two 3/1 red Elemental creature tokens with haste. Exile them at the beginning of the next end step.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 2, CreatureTokens: 2},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("3/1 red Elemental with haste")), Count: 2}},
					}, ctx.ItemTail(chandraFlamecallerExile))
				},
			},
			{
				Label: "0: Discard all the cards in your hand, then draw that many cards plus one.",
				Cost:  effects.LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					n, err := effects.DiscardWholeHand(g, item.Controller)
					if err != nil {
						return err
					}
					return effects.DrawCards{N: n + 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−X: Chandra deals X damage to each creature.",
				Cost:    effects.LoyaltyMinusX(),
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if ctx.X() <= 0 {
						return nil
					}
					var ids []uuid.UUID
					for _, c := range effects.MatchingBattlefield(ctx, effects.Creature()) {
						ids = append(ids, c.InstanceID)
					}
					return effects.DealDamageToEachThen(ctx, ids, ctx.X(), effects.RecipientThen{})
				},
			},
		},
	})
}

var chandraFlamecallerExile = effects.SocB8ExileTokensAtNextEndStep("chandra-flamecaller/exile-the-elementals",
	"Chandra, Flamecaller — exile the Elementals")
