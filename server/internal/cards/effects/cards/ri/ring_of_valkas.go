package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ring of Valkas — Artifact — Equipment {2} (Commander Legends):
//
//	"Equipped creature has haste.
//	 At the beginning of your upkeep, put a +1/+1 counter on equipped creature
//	 if it's red.
//	 Equip {1}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "412f57e8-76af-4636-bab0-e0940300ee2e",
		Name:         "Ring of Valkas",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("haste")},
		Triggered: []game.TriggeredAbility{
			effects.OldB084UpkeepCounterIfEquippedIs("Ring of Valkas — put a +1/+1 counter on equipped creature if it's red", "R"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
