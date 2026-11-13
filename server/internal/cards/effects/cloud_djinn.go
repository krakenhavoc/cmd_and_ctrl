package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloud Djinn — Creature — Djinn {5}{U}, 5/4 (Weatherlight, #5629):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Flying is the printed keyword. The block restriction is Stormbound
// Geist's block rule: the Djinn may not block an attacker without flying
// (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4914d38d-ed6b-4531-b72c-06f0968c44e9",
		Name:            "Cloud Djinn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
