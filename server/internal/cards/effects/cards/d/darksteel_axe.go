package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darksteel Axe — Artifact — Equipment {1} (Scars of Mirrodin, #5577):
//
//	"Indestructible
//	 Equipped creature gets +2/+0.
//	 Equip {2}"
//
// Bonesplitter's pump on an Equipment that prints indestructible
// (CR 702.12b); the keyword is the Axe's own and is not granted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3b7ea3ac-ac0a-40aa-b743-d153e1c47d8c",
		Name:            "Darksteel Axe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{effects.PumpAttached(2, 0)},
		Activated:       []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
