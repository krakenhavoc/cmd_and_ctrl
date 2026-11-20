package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steelform Sliver — Creature — Sliver {2}{W}, 2/2 (Magic 2014, #5563):
//
//	"Sliver creatures you control get +0/+1."
//
// The modern wording names "Sliver creatures you control": the anthem includes
// this Sliver itself and leaves the Slivers across the table alone
// (TribeFilter.YoursOnly). It is a layer 7c modification, read live, so a
// Sliver that arrives later gets it at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae58e71c-b617-404f-a5dc-5e6a487f2675",
		Name:         "Steelform Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, 0, 1),
		},
	})
}
