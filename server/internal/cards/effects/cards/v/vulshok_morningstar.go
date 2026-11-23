package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulshok Morningstar — Artifact — Equipment {2} (Darksteel, #5605):
//
//	"Equipped creature gets +2/+2.
//	 Equip {2} ({2}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// Leonin Scimitar's shape: the shared attached pump and equip ability
// (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12a8adc4-927f-4314-b2ef-9c647ace68d5",
		Name:         "Vulshok Morningstar",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
