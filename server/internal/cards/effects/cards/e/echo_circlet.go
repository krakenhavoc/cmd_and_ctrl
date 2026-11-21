package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Echo Circlet — Artifact — Equipment {2}:
//
//	"Equipped creature can block an additional creature each combat.
//	 Equip {1}"
//
// CanBlockAdditional over the equipped creature (#1706), and the
// shared EquipAbility. The effect is the Equipment's, so a host that
// has lost its abilities still blocks two.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17134950-fdb3-48d0-b541-419db673f4c9",
		Name:         "Echo Circlet",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.CanBlockAdditional(effects.AttachedToSource, 1)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
