package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coalhauler Swine — Creature — Boar Beast {4}{R}{R}, 4/4 (Ravnica: City
// of Guilds, #5600):
//
//	"Whenever this creature is dealt damage, it deals that much damage to
//	 each player."
//
// Boros Reckoner's trigger (one per batch of damage, carrying the total the
// Swine was dealt; CR 120.3, CR 510.2) with Flame Rift's each-player body.
// The Swine is the source, so it deals the damage even after it died to
// what it was dealt. Every player is hit, its controller included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c00f082b-787f-4f96-8356-c10d35713bc4",
		Name:         "Coalhauler Swine",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Coalhauler Swine — it deals that much damage to each player", coalhaulerSwineHitsEachPlayer),
		},
	})
}

// coalhaulerSwineHitsEachPlayer deals the damage the Swine was dealt, read
// off the trigger's event, to each player.
func coalhaulerSwineHitsEachPlayer(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return effects.B38DamageEachPlayer(effects.NewContext(g, item), item.Trigger.Event.Amount)
}
