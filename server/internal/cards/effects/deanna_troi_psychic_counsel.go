package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deanna Troi, Psychic Counsel — {1}{W} Legendary Creature — Betazoid
// Human Advisor 2/2 (Star Trek, #4973):
//
//	"Other creatures you control enter with an additional +1/+1 counter
//	 on them."
//
// Renata's CR 614.12 entry replacement
// (b27OtherCreaturesYouControlEnterWithACounter): cast, reanimated,
// flickered and token creatures all enter with the counter, and Hardened
// Scales and Doubling Season apply to it. Deanna herself gets none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf612cbe-7af8-48fd-875e-a669e05cfb49",
		Name:         "Deanna Troi, Psychic Counsel",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b27OtherCreaturesYouControlEnterWithACounter("Deanna Troi, Psychic Counsel: enters with an additional +1/+1 counter"),
		},
	})
}
