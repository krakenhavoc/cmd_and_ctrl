package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Snake of the Golden Grove — {4}{G} Creature — Snake 4/4 (Born of the Gods):
//
//	"Tribute 3 (As this creature enters, an opponent of your choice may put three
//	 +1/+1 counters on it.)
//	 When this creature enters, if tribute wasn't paid, you gain 4 life."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef0c8193-8cbe-409e-9c2e-96b46ffcc420",
		Name:         "Snake of the Golden Grove",
		Completeness: effects.CompletenessFull,
		Replacements: effects.Tribute("Snake of the Golden Grove", 3),
		Triggered: []game.TriggeredAbility{
			effects.WhenEntersIfTributeWasntPaid("Snake of the Golden Grove — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
