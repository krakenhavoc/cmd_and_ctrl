package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c00f082b-787f-4f96-8356-c10d35713bc4",
		Name:         "Coalhauler Swine",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Coalhauler Swine — it deals that much damage to each player", coalhaulerSwineHitsEachPlayer),
		},
	})
}

// coalhaulerSwineHitsEachPlayer deals the damage the Swine was dealt, read
// off the trigger's event, to each player.
func coalhaulerSwineHitsEachPlayer(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return b38DamageEachPlayer(NewContext(g, item), item.Trigger.Event.Amount)
}
