package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vaporkin — Creature — Elemental {1}{U}, 2/1 (Theros, #5561):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Cloud Spirit's block rule: it may block an attacker only if that
// attacker has flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "169e415a-9ff6-43bb-bb68-fc49e69f95a8",
		Name:            "Vaporkin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
