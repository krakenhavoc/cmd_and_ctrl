package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystal Slipper — Artifact — Equipment {1}{R} (Throne of Eldraine, #5516):
//
//	"Equipped creature gets +1/+0 and has haste.
//	 Equip {1}"
//
// Boots of Speed's text at a red cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c2d2888-38be-41f9-aaf2-8337e87d2aed",
		Name:         "Crystal Slipper",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0), effects.GrantToAttached("haste")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
