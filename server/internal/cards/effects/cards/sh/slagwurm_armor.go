package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slagwurm Armor — Artifact — Equipment {1} (Mirrodin, #5606):
//
//	"Equipped creature gets +0/+6.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// A layer 7c boost on the equipped creature (PumpAttached), and the
// shared sorcery-speed equip row (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20b60c93-124b-42c8-93fb-63bbd1888658",
		Name:         "Slagwurm Armor",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(0, 6)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
