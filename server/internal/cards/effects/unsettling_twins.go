package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unsettling Twins — Creature — Human {3}{W}:
//
//	"When this creature enters, manifest dread."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0b2ca8c-cb5a-45f8-ae7c-63ad9ebb4afd",
		Name:         "Unsettling Twins",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Unsettling Twins — manifest dread", Do(ManifestDread{})),
		},
	})
}
