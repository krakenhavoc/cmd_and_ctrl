package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turntimber Ascetic — Creature — Giant Cleric {4}{G}, 5/4 (Zendikar
// Rising, #5508):
//
//	"When this creature enters, you gain 3 life."
//
// Dawnhart Rejuvenator's enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90b16d52-4cf4-4788-a0d7-264d1bcd17f0",
		Name:         "Turntimber Ascetic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Turntimber Ascetic — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
