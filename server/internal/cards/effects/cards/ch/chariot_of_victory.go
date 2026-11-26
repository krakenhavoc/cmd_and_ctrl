package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chariot of Victory — Artifact — Equipment {3} (Journey into Nyx, S82 OLD,
// #5558):
//
//	"Equipped creature has first strike, trample, and haste.
//	 Equip {1}"
//
// Bladed Pinions' keyword grant to the attached creature with three keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a40339d-7821-4f37-a957-1a6ebdd993e0",
		Name:         "Chariot of Victory",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("first strike", "trample", "haste")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
