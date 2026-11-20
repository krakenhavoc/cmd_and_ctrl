package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreadwurm — Creature — Wurm Horror {4}{B}, 5/4 (Zendikar Rising, #5508):
//
//	"Landfall — Whenever a land you control enters, this creature gains
//	 indestructible until end of turn."
//
// Landfall is an ability word (CR 207.2c). The grant is Bark-Knuckle
// Boxer's thisCreatureUntilEOT with no pump, read at resolution and
// doing nothing if the Wurm has left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9feb7600-f844-4573-bcb7-d6b34473517b",
		Name:         "Dreadwurm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Dreadwurm — this creature gains indestructible until end of turn (landfall)",
				thisCreatureUntilEOT("Dreadwurm — indestructible", 0, 0, "indestructible")),
		},
	})
}
