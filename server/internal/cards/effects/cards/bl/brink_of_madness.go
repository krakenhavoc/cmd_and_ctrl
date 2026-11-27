package cardsbl

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brink of Madness — Enchantment {2}{B}{B}:
//
//	"At the beginning of your upkeep, if you have no cards in hand,
//	 sacrifice this enchantment and target opponent discards their hand."
//
// An empty hand is the intervening "if" (CR 603.4): checked as the ability
// triggers and as it resolves. The target is chosen as it goes on the stack;
// an opponent who is no longer a legal target leaves the ability countered on
// resolution (CR 608.2b), so the enchantment stays.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3209ad08-5482-4dd4-a02e-1dbe99ce6b54",
		Name:         "Brink of Madness",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.IfVentureCond(func(g *game.Game, you uuid.UUID) bool {
				p := g.PlayerByIDForEffect(you)
				return p != nil && p.Hand.Size() == 0
			}, effects.AtYourUpkeep("Brink of Madness — sacrifice it and target opponent discards their hand", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				opp, ok := ctx.ClauseTarget(0)
				if !ok || opp.Kind != game.TargetPlayer {
					return nil
				}
				if err := effects.SacrificeThisIfStillHere(ctx); err != nil {
					return err
				}
				_, err := effects.DiscardWholeHand(g, opp.ID)
				return err
			})), effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
