package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battleground Geist — Creature — Spirit {4}{U}, 3/3 (Innistrad Remastered,
// S82 GAP, #5948):
//
//	"Flying
//	 Other Spirit creatures you control get +1/+0."
//
// Drogskol Captain's lord scoped to the OTHER Spirits its controller
// controls (the shared tribal anthem, layer 7c), so a Spirit token made
// later gets the bonus the moment it exists and the Geist itself doesn't.
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
