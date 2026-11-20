package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kite Shield — Artifact — Equipment {0} (Magic 2012, #5573):
//
//	"Equipped creature gets +0/+3.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// Accorder's Shield without the vigilance: PumpAttached and the shared
// equip row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24f9c64a-6ce6-41cd-948e-11554d7f17fb",
		Name:         "Kite Shield",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(0, 3)},
		Activated:    []ActivatedAbility{EquipAbility("{3}")},
	})
}
