package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Staff of the Mind Magus — Artifact {3} (Magic 2014, #5563):
//
//	"Whenever you cast a blue spell or an Island you control enters, you gain 1 life."
//
// One ability with two trigger conditions (m14StaffOfTheMagus): casting a
// blue spell, and a Island entering under its controller's control. Each
// is its own trigger, and each gains 1 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1466687f-cfe1-4e6e-ad8e-dca5e8e4c06d",
		Name:         "Staff of the Mind Magus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			m14StaffOfTheMagus("Staff of the Mind Magus", "U", "Island"),
		},
	})
}
