package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Lord of the Eagles — Legendary Creature — Bird Noble {7}{U}{U}, 8/8
// (The Hobbit):
//
//	"Flash
//	 This spell costs {X} less to cast, where X is the total power of
//	 creatures you control with flying.
//	 Flying"
//
// Ghalta's self cost reduction with a keyword on the creatures counted
// (TotalPowerOfMatchingCreaturesYouControl): current power, so counters and
// anthems count, and a creature that has flying from any source counts
// (game.HasKeyword reads the effective keywords). The reduction spends
// against generic mana only, so the {U}{U} always stays (CR 601.2f), and a
// total of zero or less reduces nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b38f314-d997-4c2a-b4c9-40dbe780993d",
		Name:            "The Lord of the Eagles",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(TotalPowerOfMatchingCreaturesYouControl(func(c game.Card) bool { return game.HasKeyword(&c, "flying") }),
				"This spell costs {X} less to cast, where X is the total power of creatures you control with flying."),
		},
	})
}
