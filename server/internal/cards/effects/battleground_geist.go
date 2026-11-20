package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battleground Geist — Creature — Spirit {4}{U}, 3/3 (Innistrad, #5572):
//
//	"Flying
//	 Other Spirit creatures you control get +1/+0."
//
// Drogskol Captain's tribal anthem (TribalAnthem over the other Spirits
// you control) at +1/+0, with printed flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2450d35b-7d87-4885-b396-f8e1d3ac0b22",
		Name:            "Battleground Geist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Spirit"}, Others: true, YoursOnly: true}, 1, 0),
		},
	})
}
