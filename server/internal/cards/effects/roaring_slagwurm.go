package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roaring Slagwurm — Creature — Wurm {5}{G}{G} (Darksteel, #5605):
//
//	"Whenever this creature attacks, tap all artifacts."
//
// An attack trigger (CR 508.3a) whose body is Metal Fatigue's: every
// artifact on the battlefield, its controller's included, is tapped as it
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "856d8eb4-3834-49f8-9696-579948590fd0",
		Name:         "Roaring Slagwurm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Roaring Slagwurm — tap all artifacts",
				func(g *game.Game, item *game.StackItem) error {
					return dstB2TapAllArtifacts(NewContext(g, item))
				}),
		},
	})
}
