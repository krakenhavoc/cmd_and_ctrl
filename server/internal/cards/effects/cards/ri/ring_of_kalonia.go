package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ring of Kalonia — Artifact — Equipment {2} (Magic 2013, S82 OLD, #5568):
//
//	"Equipped creature has trample.
//	 At the beginning of your upkeep, put a +1/+1 counter on equipped creature
//	 if it's green.
//	 Equip {1}"
//
// Ring of Thune's shape with trample for vigilance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4344e5e-148b-4df9-88e4-ca91afcc32e0",
		Name:         "Ring of Kalonia",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("trample")},
		Triggered: []game.TriggeredAbility{
			effects.OldB084UpkeepCounterIfEquippedIs("Ring of Kalonia — put a +1/+1 counter on equipped creature if it's green", "G"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
