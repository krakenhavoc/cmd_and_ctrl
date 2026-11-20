package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cathar's Shield — Artifact — Equipment for {0}:
//
//	"Equipped creature gets +0/+3 and has vigilance.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// A defensive stat boost plus a keyword grant — PumpAttached and
// GrantToAttached composed the way every other vanilla Equipment in
// the catalog is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b270d091-ea4c-4dfd-9c50-48b45bbc396e",
		Name:         "Cathar's Shield",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 3),
			effects.GrantToAttached("vigilance"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{3}"),
		},
	})
}
