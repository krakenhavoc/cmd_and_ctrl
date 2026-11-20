package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gallows Warden — Creature — Spirit {4}{W}, 3/3 (Innistrad, #5572):
//
//	"Flying
//	 Other Spirit creatures you control get +0/+1."
//
// Drogskol Captain's tribal anthem (TribalAnthem over the other Spirits
// you control) at +0/+1, with printed flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "06157e34-dbe1-4936-a718-3b1d70726794",
		Name:            "Gallows Warden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Spirit"}, Others: true, YoursOnly: true}, 0, 1),
		},
	})
}
