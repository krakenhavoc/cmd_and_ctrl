package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dinrova Horror — Creature — Horror {4}{U}{B}, 4/4 (Jumpstart, #5510):
//
//	"When this creature enters, return target permanent to its owner's hand,
//	 then that player discards a card."
//
// Riftwing Cloudskate's bounce, then the permanent's owner (read before it
// moves) discards a card of their choice (CR 701.9a). A token returned this
// way ceases to exist, and its owner still discards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1eb7a92-3055-426e-9716-cc6d2a69f0af",
		Name:         "Dinrova Horror",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Dinrova Horror — return target permanent to its owner's hand, then that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					c, ok := g.LookupCardForEffect(id)
					if !ok {
						return nil
					}
					owner := c.Owner
					if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
						return err
					}
					return discardOneForDinrova(g, item, owner)
				}), effects.TargetPermanent("target permanent")),
		},
	})
}

func discardOneForDinrova(g *game.Game, item *game.StackItem, player uuid.UUID) error {
	if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: item.SourceCardID, N: 1})
	return nil
}
