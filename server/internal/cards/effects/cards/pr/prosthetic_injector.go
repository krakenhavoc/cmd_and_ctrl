package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prosthetic Injector — Artifact — Equipment {1}
// (Phyrexia: All Will Be One, #5479):
//
//	"Equipped creature gets +0/+2 and has toxic 1.
//	 Equip {1}"
//
// The toxic is cumulative (CR 702.164b) with any the creature has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8026533d-c0e7-48b7-9e1a-9a370899aa57",
		Name:         "Prosthetic Injector",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 2),
			effects.GrantToAttached("toxic 1"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
