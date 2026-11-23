package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aeronaut's Wings — Artifact — Equipment {2} (The Brothers' War):
//
//	"Equipped creature gets +1/+0 and has flying.
//	 Equip {2}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3bad2e2-1024-4261-b3d7-82a6998416fd",
		Name:         "Aeronaut's Wings",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("flying"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
