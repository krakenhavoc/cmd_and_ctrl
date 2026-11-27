package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saprazzan Bailiff — Creature — Merfolk {3}{U}{U}, 2/2 (Mercadian Masques,
// S82 OLD, #5619):
//
//	"When this creature enters, exile all artifact and enchantment cards from
//	 all graveyards.
//	 When this creature leaves the battlefield, return all artifact and
//	 enchantment cards from all graveyards to their owners' hands."
//
// Two triggers. Each reads the graveyards as it resolves, every seat's, and
// moves the cards as one batch. The cards exiled on entering are gone for good:
// the leave trigger returns whatever artifact and enchantment cards are in the
// graveyards then, never the exiled ones. A card that is both an artifact and an
// enchantment is one card.
//
// No simplification.
func init() {
	artifactOrEnchantmentCard := func(c game.Card) bool { return c.IsArtifact() || c.IsEnchantment() }
	effects.Register(effects.Spec{
		OracleID:     "1f861517-ac53-4218-b62e-35e971098b6d",
		Name:         "Saprazzan Bailiff",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Saprazzan Bailiff — exile all artifact and enchantment cards from all graveyards",
				func(g *game.Game, item *game.StackItem) error {
					ids := effects.AllGraveyardsCardIDs(effects.NewContext(g, item), artifactOrEnchantmentCard)
					if len(ids) == 0 {
						return nil
					}
					g.ExileCardsForEffect(ids)
					return nil
				}),
			effects.WhenThisLeaves("Saprazzan Bailiff — return all artifact and enchantment cards from all graveyards to their owners' hands",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, id := range effects.AllGraveyardsCardIDs(ctx, artifactOrEnchantmentCard) {
						if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
