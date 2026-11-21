package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "51024f9a-6c6e-4555-a156-92b32255a0c9",
		Name:         "Greatsword",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 0)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
