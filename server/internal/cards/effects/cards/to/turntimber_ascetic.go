package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turntimber Ascetic — Creature — Giant Cleric {4}{G}, 5/4 (Zendikar
// Rising, #5508):
//
//	"When this creature enters, you gain 3 life."
//
// Dawnhart Rejuvenator's enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90b16d52-4cf4-4788-a0d7-264d1bcd17f0",
		Name:         "Turntimber Ascetic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Turntimber Ascetic — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
