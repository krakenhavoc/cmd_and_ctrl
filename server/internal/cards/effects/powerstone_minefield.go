package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Powerstone Minefield — Enchantment {2}{R}{W} (Apocalypse, #5614):
//
//	"Whenever a creature attacks or blocks, this enchantment deals 2
//	 damage to it."
//
// Two triggers over every player's creatures, the Minefield's
// controller's included: one per attacking creature as attackers are
// declared (aCreatureAttacked, CR 508.1m) and one per blocking creature
// as blockers are declared (aCreatureBlocked, CR 509.1i), a creature
// that blocks several attackers blocking once. "It" is the creature the
// event names; one that has left the battlefield by resolution is dealt
// nothing. A creature put onto the battlefield attacking never attacked
// (CR 508.4), so it is not dealt damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a8a67ad-e4ff-4b55-9976-a2f13cc9b41b",
		Name:         "Powerstone Minefield",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, aCreatureAttacked,
				"Powerstone Minefield — deals 2 damage to the attacking creature", powerstoneMinefieldDamage),
			On(game.EventBlock, aCreatureBlocked,
				"Powerstone Minefield — deals 2 damage to the blocking creature", powerstoneMinefieldDamage),
		},
	})
}

// powerstoneMinefieldDamage deals 2 damage from the Minefield to the
// creature that attacked or blocked, while it is still on the
// battlefield.
func powerstoneMinefieldDamage(g *game.Game, item *game.StackItem) error {
	it := item.Trigger.Event.CardID
	if !onBattlefield(g, it) {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: it, Amount: 2}.Apply(NewContext(g, item))
}
