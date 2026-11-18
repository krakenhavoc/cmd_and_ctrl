package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Supreme Phantom — Creature — Spirit {1}{U}, 1/3:
//
//	"Flying
//	 Other Spirits you control get +1/+1."
//
// Legion Lieutenant's tribal anthem over Spirits.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9c5f4d02-eedb-4c6c-9f13-1b7a45382483",
		Name:            "Supreme Phantom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Spirit"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
