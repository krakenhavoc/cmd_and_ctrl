package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thraben Watcher — Creature — Angel {2}{W}{W}, 2/2 (Foundations Commander):
//
//	"Flying, vigilance
//	 Other nontoken creatures you control get +1/+1 and have vigilance."
//
// Two statics over one AppliesTo, the way Intangible Virtue and Death Baron
// split their sentences: the +1/+1 is layer 7c, the vigilance grant layer 6.
// "Other" is object identity; "nontoken" reads the token flag, so a token
// the Watcher's controller makes is neither pumped nor given vigilance.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "53cd9d9f-c1a7-404e-9934-4542415e3431",
		Name:            "Thraben Watcher",
		PrintedKeywords: []string{"flying", "vigilance"},
		Completeness:    CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(cmdrB3OtherNontokenCreaturesYouControl, 1, 1),
			b16GrantKeywords(cmdrB3OtherNontokenCreaturesYouControl, "vigilance"),
		},
	})
}
