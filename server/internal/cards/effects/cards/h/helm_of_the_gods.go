package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Helm of the Gods — Artifact — Equipment {1} (Magic Origins, S82 older sets):
//
//	"Equipped creature gets +1/+1 for each enchantment you control.
//	 Equip {1}"
//
// Ethereal Armor's count on an Equipment: a layer 7c pump on the equipped
// creature, re-read on every layer recompute (CR 613.4c), counting the
// enchantments its controller controls through the layered view (the Helm's own
// controller, "you"). Unattached it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52804094-47b9-40b9-999d-7c050d220e63",
		Name:         "Helm of the Gods",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttachedPer(1, 1, effects.EnchantmentsControlledBy)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
