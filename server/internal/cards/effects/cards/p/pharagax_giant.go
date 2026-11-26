package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pharagax Giant — {4}{R} Creature — Giant 3/3 (Born of the Gods):
//
//	"Tribute 2 (As this creature enters, an opponent of your choice may put two
//	 +1/+1 counters on it.)
//	 When this creature enters, if tribute wasn't paid, this creature deals 5
//	 damage to each opponent."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69a134f2-c0c4-496f-b354-8325437af8ca",
		Name:         "Pharagax Giant",
		Completeness: effects.CompletenessFull,
		Replacements: effects.Tribute("Pharagax Giant", 2),
		Triggered: []game.TriggeredAbility{
			effects.WhenEntersIfTributeWasntPaid("Pharagax Giant — it deals 5 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 5})),
		},
	})
}
