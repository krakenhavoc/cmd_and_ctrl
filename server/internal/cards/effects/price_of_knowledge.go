package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Price of Knowledge — Enchantment {6}{B}:
//
//	"Players have no maximum hand size.
//	 At the beginning of each opponent's upkeep, this enchantment deals
//	 damage to that player equal to the number of cards in that
//	 player's hand."
//
// "Players have no maximum hand size" reaches every player, its
// controller included (ADR 0113 §3, #2074), and is folded in CR
// 613.11's timestamp order: an opponent under an earlier Jin-Gitaxias
// has no maximum, and one under a later one is back at zero. The
// damage counts that player's hand when the trigger resolves (the
// 2013-10-17 ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c586aa7-7a61-464f-abba-b33f9a525f0e",
		Name:         "Price of Knowledge",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{PlayersHaveNoMaxHandSize()},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, ByAnOpponent, "Price of Knowledge — damage equal to that player's hand size", func(g *game.Game, item *game.StackItem) error {
				p := g.PlayerByIDForEffect(triggeringActor(item))
				if p == nil || p.ID == uuid.Nil || p.Hand == nil || p.Hand.Size() == 0 {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: p.ID, Amount: p.Hand.Size()}.Apply(NewContext(g, item))
			}),
		},
	})
}
