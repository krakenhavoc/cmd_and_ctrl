package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Predatory Sliver — Creature — Sliver {1}{G}, 1/1 (Magic 2014, #5563):
//
//	"Sliver creatures you control get +1/+1."
//
// The modern wording names "Sliver creatures you control": the anthem includes
// this Sliver itself and leaves the Slivers across the table alone
// (TribeFilter.YoursOnly). It is a layer 7c modification, read live, so a
// Sliver that arrives later gets it at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b589e70-9c1d-48dd-a597-479959e692bf",
		Name:         "Predatory Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, 1, 1),
		},
	})
}
