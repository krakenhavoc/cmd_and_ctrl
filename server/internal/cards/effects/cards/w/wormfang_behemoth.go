package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wormfang Behemoth — Creature — Nightmare Fish Beast {3}{U}{U}, 5/5
// (Judgment):
//
//	"When this creature enters, exile all cards from your hand.
//	 When this creature leaves the battlefield, return the exiled cards to
//	 their owner's hand."
//
// The original linked pair of the Wormfang cycle: the enters trigger exiles
// the whole hand, each card linked to this object (CR 607.2a), and the leaves
// trigger returns every card still linked to the object that left to its
// owner's hand (S82OldB627ReturnLinkedExiledCardsToHand). A Behemoth
// that left before its enters trigger resolved has nothing to return, so what
// it exiles afterwards stays exiled.
//
// No simplification.
func init() {
	const name = "Wormfang Behemoth"
	effects.Register(effects.Spec{
		OracleID:     "77a000de-76d9-44f8-9789-875aa9f576e7",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — exile all cards from your hand", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				p := g.PlayerByIDForEffect(ctx.Controller())
				if p == nil || p.Hand == nil {
					return nil
				}
				ids := make([]uuid.UUID, 0, p.Hand.Size())
				for _, c := range p.Hand.Cards {
					ids = append(ids, c.InstanceID)
				}
				for _, id := range ids {
					if err := (effects.ExileTarget{Target: id, Linked: true}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.S82OldB627ReturnLinkedExiledCardsToHand(name + " — return the exiled cards to their owner's hand"),
		},
	})
}
