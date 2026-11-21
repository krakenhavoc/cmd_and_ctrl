package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Copper Carapace — Artifact — Equipment {1} (Mirrodin Besieged,
// #5576):
//
//	"Equipped creature gets +2/+2 and can't block.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// A layer 7c pump (PumpAttached) and Sluggishness's can't-block
// restriction on the host (RestrictAttached), both read live off the
// attachment, with the equip ability (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ffbff60-a59d-4f71-864e-3c16cb61e43b",
		Name:         "Copper Carapace",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.RestrictAttached(game.CantBlock),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
