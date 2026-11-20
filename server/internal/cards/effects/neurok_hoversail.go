package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neurok Hoversail — Artifact — Equipment {1} (Mirrodin, #5606):
//
//	"Equipped creature has flying.
//	 Equip {2} ({2}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// Cliffhaven Kitesail's grant without the enters trigger: flying on the
// equipped creature (GrantToAttached, layer 6), and the shared
// sorcery-speed equip row (CR 702.6a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3cc02a23-93b7-445a-9c3e-0e4942ef927b",
		Name:         "Neurok Hoversail",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("flying")},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
