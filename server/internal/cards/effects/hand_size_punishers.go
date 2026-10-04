package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hand_size_punishers.go — the upkeep cards that read a player's hand
// size: Iron Maiden and Viseling ("X damage, where X is the number of
// cards in their hand minus 4") and The Rack ("3 minus the number of
// cards in their hand"), the shape Storm World already had. They only
// READ a hand, so none of them needs the maximum-hand-size seam.
// Append-only.

// cardsInHand is the number of cards in `player`'s hand, zero for a
// seat with no hand zone.
func cardsInHand(g *game.Game, player uuid.UUID) int {
	if p := g.PlayerByIDForEffect(player); p != nil && p.Hand != nil {
		return p.Hand.Size()
	}
	return 0
}

// damageUpkeepPlayerByHand is "this deals X damage to that player,
// where X is <a function of the number of cards in their hand>": the
// player is the upkeep's player, read off the trigger's event, and X
// is counted as the trigger resolves (CR 608.2h). A result of zero or
// less deals nothing (CR 107.1b).
func damageUpkeepPlayerByHand(amount func(hand int) int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		player := NewContext(g, item).Trigger().Event.Actor
		if player == uuid.Nil {
			return nil
		}
		return DealDamage{
			Source: item.SourceCardID,
			Target: player,
			Amount: amount(cardsInHand(g, player)),
		}.Apply(NewContext(g, item))
	}
}

// handMinusFour is Iron Maiden's and Viseling's X.
func handMinusFour(hand int) int { return hand - 4 }
