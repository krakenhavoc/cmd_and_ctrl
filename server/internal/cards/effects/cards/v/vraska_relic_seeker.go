package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vraska, Relic Seeker — Legendary Planeswalker — Vraska {4}{B}{G},
// loyalty 6 (Ixalan, #5532):
//
//	"+2: Create a 2/2 black Pirate creature token with menace.
//	 −3: Destroy target artifact, creature, or enchantment. Create a
//	     Treasure token.
//	 −10: Target player's life total becomes 1."
//
// The −3 does nothing at all when its one target is illegal as it resolves
// (CR 608.2b), the Treasure included; with a legal target it destroys and
// then makes the Treasure, even if the permanent was indestructible. The −10
// sets the life total (CR 118.5): a player already at 1 or less loses or
// gains nothing, and one above 1 loses the difference as life loss.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3c037103-5af0-40a8-bbbe-7f3f5958b6e9",
		Name:            "Vraska, Relic Seeker",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 6,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Create a 2/2 black Pirate creature token with menace.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Pirate with menace"), N: 1}),
			},
			{
				Label:   "−3: Destroy target artifact, creature, or enchantment. Create a Treasure token.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target artifact, creature, or enchantment", effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment())),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if len(ctx.LegalTargets()) == 0 {
						return nil
					}
					if err := effects.DestroyTheFirstLegalTarget(g, item); err != nil {
						return err
					}
					return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "−10: Target player's life total becomes 1.",
				Cost:    effects.LoyaltyCost(-10),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return effects.B31LifeBecomes(ctx, t.ID, 1)
						}
					}
					return nil
				},
			},
		},
	})
}
