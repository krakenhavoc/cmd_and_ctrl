package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mortician Beetle — Creature — Insect {B}, 1/1 (Rise of the Eldrazi, #5581):
//
//	"Whenever a player sacrifices a creature, you may put a +1/+1 counter on
//	 this creature."
//
// The sacrifice event fires while the permanent is still on the battlefield,
// so whether it was a creature is read as it last stood (CR 603.10a). Any
// player's sacrifice counts, the Beetle's controller's included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc5c4a38-a415-466a-a7ea-dc04eee7f691",
		Name:         "Mortician Beetle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventSacrifice, roeB1APlayerSacrificedACreature,
				"Mortician Beetle — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				"Mortician Beetle — put a +1/+1 counter on it?"),
		},
	})
}
