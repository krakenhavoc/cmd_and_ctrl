package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cobbled Wings — Artifact — Equipment {2} (Innistrad, #5572):
//
//	"Equipped creature has flying.
//	 Equip {1} ({1}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// Bladed Pinions' layer-6 grant (GrantToAttached) with flying alone, and
// the shared sorcery-speed Equip row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d8682f3-9ef3-4aa7-9ea6-8a2ce09bff6f",
		Name:         "Cobbled Wings",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
