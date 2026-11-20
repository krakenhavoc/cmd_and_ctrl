package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Staff of the Death Magus — Artifact {3} (Magic 2014, #5563):
//
//	"Whenever you cast a black spell or a Swamp you control enters, you gain 1 life."
//
// One ability with two trigger conditions (m14StaffOfTheMagus): casting a
// black spell, and a Swamp entering under its controller's control. Each
// is its own trigger, and each gains 1 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3fda16e-fd56-434e-91ac-561ae0482f43",
		Name:         "Staff of the Death Magus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			m14StaffOfTheMagus("Staff of the Death Magus", "B", "Swamp"),
		},
	})
}
