package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loxodon Punisher — Creature — Elephant Soldier {3}{W}, 2/2 (Mirrodin, #5606):
//
//	"This creature gets +2/+2 for each Equipment attached to it."
//
// A layer 7c count of the Equipment attached to the Punisher itself,
// whoever controls them (mrdB5PerEquipmentOnThis), read live, so an
// Equipment that moves away takes its +2/+2 with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc3869fd-1441-4d5e-9af5-f02857ed634a",
		Name:         "Loxodon Punisher",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{mrdB5PerEquipmentOnThis(2)},
	})
}
