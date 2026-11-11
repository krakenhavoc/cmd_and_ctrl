package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winged Boots — Artifact — Equipment {1}{U} (Outlaws of Thunder Junction
// Commander):
//
//	"Equipped creature has flying and ward {4}.
//	 Equip {1}"
//
// Flying is the shared keyword grant; ward is a triggered ability (CR 702.21a)
// the Equipment carries for its current host, the Lavaspur Boots shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a178a789-1f5f-46ca-a90b-574bd236a671",
		Name:         "Winged Boots",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			WardAttached(WardMana("{4}"), "Winged Boots — ward {4}"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
