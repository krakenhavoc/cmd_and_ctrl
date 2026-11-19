package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gorgon Flail — Artifact — Equipment {2} (Foundations Jumpstart):
//
//	"Equipped creature gets +1/+1 and has deathtouch.
//	 Equip {2}"
//
// Bonesplitter's shape with a deathtouch grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91dab3e2-039e-4db9-94e3-1146bced18ed",
		Name:         "Gorgon Flail",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1), GrantToAttached("deathtouch")},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
