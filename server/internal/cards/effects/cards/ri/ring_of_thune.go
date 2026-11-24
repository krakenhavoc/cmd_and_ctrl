package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ring of Thune — Artifact — Equipment {2} (Commander Legends):
//
//	"Equipped creature has vigilance.
//	 At the beginning of your upkeep, put a +1/+1 counter on equipped creature
//	 if it's white.
//	 Equip {1}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f72e8ccb-a6fb-4161-bfa1-69c94cc5d49a",
		Name:         "Ring of Thune",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("vigilance")},
		Triggered: []game.TriggeredAbility{
			effects.OldB084UpkeepCounterIfEquippedIs("Ring of Thune — put a +1/+1 counter on equipped creature if it's white", "W"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
