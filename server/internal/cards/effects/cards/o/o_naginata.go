package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// O-Naginata — Artifact — Equipment {1} (Saviors of Kamigawa, S82 OLD b593):
//
//	"This Equipment can be attached only to a creature with power 3 or greater.
//	 Equipped creature gets +3/+0 and has trample.
//	 Equip {2}"
//
// Gate Smasher's shape with a power bound (CR 301.5b, 301.5c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5641c831-5254-4ef2-ba48-328d86007479",
		Name:         "O-Naginata",
		Completeness: effects.CompletenessFull,
		AttachOnlyTo: effects.AttachedOnlyTo("a creature with power 3 or greater", effects.PowerGE(3)),
		Static:       []game.StaticAbility{effects.PumpAttached(3, 0), effects.GrantToAttached("trample")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
