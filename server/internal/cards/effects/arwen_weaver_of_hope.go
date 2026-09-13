package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arwen, Weaver of Hope — Legendary Creature — Elf Noble {1}{G}{G},
// 2/1 (EDHREC rank 1865):
//
//	"Each other creature you control enters with a number of
//	 additional +1/+1 counters on it equal to Arwen's toughness."
//
// The counters deck's Master Biomancer. A CR 614 replacement on the
// entry of every other creature under Arwen's controller's control —
// cast, reanimated, fetched or flickered — adding Arwen's toughness
// as it is at that moment, counters and anthems included, so a
// pumped Arwen hands out more. The counters are on the creature
// before any ETB trigger or state check sees it, and a "whenever
// you put counters" payoff does not see them, the Hangarback Walker
// posture.
//
// Sandbox simplification, declared: creature TOKENS get nothing.
// Token creation skips the CR 614 zone-move pipeline (the Urabrask
// the Hidden gap), so no entry replacement sees a token. Weaker
// than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "c0891157-f3dd-4866-866f-fd13e9be9633",
		Name:         "Arwen, Weaver of Hope",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Creature tokens you create don't get the extra +1/+1 counters — only creature cards entering the battlefield do."},
		Replacements: []game.ReplacementEffect{
			b17OtherCreaturesYouControlEnterWithCounters("Arwen, Weaver of Hope: enters with +1/+1 counters equal to Arwen's toughness"),
		},
	})
}
