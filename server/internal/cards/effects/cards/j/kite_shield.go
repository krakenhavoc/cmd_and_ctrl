package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "24f9c64a-6ce6-41cd-948e-11554d7f17fb",
		Name:         "Kite Shield",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(0, 3)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
