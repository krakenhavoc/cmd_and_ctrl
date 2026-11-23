package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vexing Drone — Artifact Creature — Borg {1}{W}, 2/2 (Star Trek):
//
//	When this creature enters, tap target creature an opponent controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e90e417-65ca-48bc-a32d-b93976519fe9",
		Name:         "Vexing Drone",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Vexing Drone — tap target creature an opponent controls", effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
