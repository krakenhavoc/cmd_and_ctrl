package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vexing Drone — Artifact Creature — Borg {1}{W}, 2/2 (Star Trek):
//
//	When this creature enters, tap target creature an opponent controls.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e90e417-65ca-48bc-a32d-b93976519fe9",
		Name:         "Vexing Drone",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Vexing Drone — tap target creature an opponent controls", tapTheFirstLegalTarget),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
