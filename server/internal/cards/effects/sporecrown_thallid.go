package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sporecrown Thallid — Creature — Fungus {1}{G}, 2/2 (Foundations
// Jumpstart):
//
//	"Each other creature you control that's a Fungus or Saproling gets
//	 +1/+1."
//
// A tribal anthem over two creature types, excluding itself and limited
// to your creatures (Anowon's shape with two tribes).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0c2ca8a-c5d4-4684-bacb-46c04f47c276",
		Name:         "Sporecrown Thallid",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Fungus", "Saproling"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
