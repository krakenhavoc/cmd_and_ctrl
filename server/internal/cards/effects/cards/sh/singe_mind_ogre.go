package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Singe-Mind Ogre — Creature — Ogre Mutant {2}{B}{R}, 3/2 (Alara Reborn,
// #5586):
//
//	"When this creature enters, target player reveals a card at random from
//	 their hand, then loses life equal to that card's mana value."
//
// Hired Torturer's keyed random reveal from the target's hand, then the life
// loss for the revealed card's mana value (a land is 0, {X} is 0 in a hand);
// an empty hand reveals nothing and loses nothing. A target that is no longer
// legal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0f26f84-9607-4411-bca5-681ba4e57756",
		Name:         "Singe-Mind Ogre",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Singe-Mind Ogre — target player reveals a card at random from their hand, then loses life equal to its mana value",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					victim := effects.TargetedPlayer(ctx)
					if victim == uuid.Nil {
						return nil
					}
					id := effects.S82OldB406RevealRandomCardFromHand(ctx, victim, "Singe-Mind Ogre")
					if id == uuid.Nil {
						return nil
					}
					c, ok := g.LookupCardForEffect(id)
					if !ok {
						return nil
					}
					mv, _ := g.ManaValueForEffect(c)
					if mv <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(ctx.Source(), victim, -mv)
				}),
				effects.TargetPlayer("target player")),
		},
	})
}
