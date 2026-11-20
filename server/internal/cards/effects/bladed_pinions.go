package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bladed Pinions — Artifact — Equipment {2} (Scars of Mirrodin, #5577):
//
//	"Equipped creature has flying and first strike.
//	 Equip {2}"
//
// Whispersilk Cloak's keyword grant with two keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8cdc08a-975b-475f-8e30-85d61f6f3a9b",
		Name:         "Bladed Pinions",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("flying", "first strike")},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
