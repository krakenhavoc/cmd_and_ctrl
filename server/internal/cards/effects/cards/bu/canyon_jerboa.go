package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canyon Jerboa — Creature — Mouse {2}{W}, 1/2 (Zendikar Rising, #5508):
//
//	"Landfall — Whenever a land you control enters, creatures you control
//	 get +1/+1 until end of turn."
//
// Landfall is an ability word (CR 207.2c). The team pump is Timid
// Shieldbearer's: the creatures are locked in as the trigger resolves
// (CR 611.2c), so one that arrives later this turn gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "254d4918-de40-4ac4-b542-f5c67a5e21eb",
		Name:         "Canyon Jerboa",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Canyon Jerboa — creatures you control get +1/+1 until end of turn (landfall)",
				effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
					Label: "Canyon Jerboa — +1/+1"})),
		},
	})
}
