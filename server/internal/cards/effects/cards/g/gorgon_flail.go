package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gorgon Flail — Artifact — Equipment {2} (Foundations Jumpstart):
//
//	"Equipped creature gets +1/+1 and has deathtouch.
//	 Equip {2}"
//
// Bonesplitter's shape with a deathtouch grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91dab3e2-039e-4db9-94e3-1146bced18ed",
		Name:         "Gorgon Flail",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1), effects.GrantToAttached("deathtouch")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
