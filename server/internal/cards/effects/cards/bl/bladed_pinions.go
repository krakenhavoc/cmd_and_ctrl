package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bladed Pinions — Artifact — Equipment {2} (Scars of Mirrodin, #5577):
//
//	"Equipped creature has flying and first strike.
//	 Equip {2}"
//
// Whispersilk Cloak's keyword grant with two keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8cdc08a-975b-475f-8e30-85d61f6f3a9b",
		Name:         "Bladed Pinions",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("flying", "first strike")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
