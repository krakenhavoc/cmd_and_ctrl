package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veteran Swordsmith — Creature — Human Soldier {2}{W}, 3/2 (Magic 2010, #5585):
//
//	"Other Soldier creatures you control get +1/+0."
//
// Yotian Tactician's lord with +1/+0: the other Soldiers you control, and
// not an opponent's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cc71b6c4-09d4-423b-a134-2f2f2fd7b533",
		Name:         "Veteran Swordsmith",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Soldier"}, Others: true, YoursOnly: true}, 1, 0),
		},
	})
}
