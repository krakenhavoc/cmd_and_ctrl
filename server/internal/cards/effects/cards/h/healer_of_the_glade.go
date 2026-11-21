package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Healer of the Glade — Creature — Elemental {G}, 1/2 (Core Set 2020, #5518):
//
//	"When this creature enters, you gain 3 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2701b912-b672-4764-97f7-51c444f754f2",
		Name:         "Healer of the Glade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Healer of the Glade — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
