package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warlord's Axe — Artifact — Equipment {3} (Magic 2011, #5579):
//
//	"Equipped creature gets +3/+1.
//	 Equip {4}"
//
// Conductive Machete's static at +3/+1 and its equip {4}, a sorcery-speed
// attach to a creature its controller controls (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebe43cc6-42fe-44f6-b198-ef0212bf82cb",
		Name:         "Warlord's Axe",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 1)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
