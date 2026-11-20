package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veteran Armorsmith — Creature — Human Soldier {W}{W}, 2/3 (Magic 2010, #5585):
//
//	"Other Soldier creatures you control get +0/+1."
//
// Yotian Tactician's lord with +0/+1: the other Soldiers you control, and
// not an opponent's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fcb34e92-f991-401c-a494-c2919083bf9e",
		Name:         "Veteran Armorsmith",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Soldier"}, Others: true, YoursOnly: true}, 0, 1),
		},
	})
}
