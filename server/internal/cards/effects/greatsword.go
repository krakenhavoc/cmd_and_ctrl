package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greatsword — Artifact — Equipment {3} (Magic 2012, #5573):
//
//	"Equipped creature gets +3/+0.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// PumpAttached for the bonus and the shared equip row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "51024f9a-6c6e-4555-a156-92b32255a0c9",
		Name:         "Greatsword",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 0)},
		Activated:    []ActivatedAbility{EquipAbility("{3}")},
	})
}
