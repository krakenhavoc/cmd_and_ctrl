package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm World — World Enchantment {R}:
//
//	"At the beginning of each player's upkeep, this enchantment deals X
//	 damage to that player, where X is 4 minus the number of cards in
//	 their hand."
//
// Every player's upkeep, its controller's included; "that player" is
// the active player the upkeep event names. X is counted as the
// trigger resolves (CR 608.2h), and a hand of four or more cards takes
// no damage (a negative amount is zero, CR 107.1b). It is damage from
// the enchantment, so prevention and lifelink-style readers see a
// source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "868f4ab2-a846-4ad0-8720-95fd234dd36b",
		Name:         "Storm World",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Storm World — deals 4 minus that player's hand size in damage to them", stormWorldDamage),
		},
	})
}

func stormWorldDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Trigger().Event.Actor
	if player == uuid.Nil {
		return nil
	}
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: player, Amount: 4 - len(p.Hand.Cards)}.Apply(ctx)
}
