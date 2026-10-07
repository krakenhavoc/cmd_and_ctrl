package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bashful Beastie — Creature — Beast {4}{G}:
//
//	"When this creature dies, manifest dread."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7aeabd9-a72c-47ee-b275-b025d7142766",
		Name:         "Bashful Beastie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Bashful Beastie — manifest dread", Do(ManifestDread{})),
		},
	})
}
