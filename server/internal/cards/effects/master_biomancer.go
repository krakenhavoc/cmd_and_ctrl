package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Master Biomancer — Creature — Elf Wizard {2}{G}{U}, 2/4 (Ravnica: Clue
// Edition, Ravnica Allegiance):
//
//	"Each other creature you control enters with a number of additional
//	 +1/+1 counters on it equal to this creature's power and as a Mutant in
//	 addition to its other types."
//
// Arwen, Weaver of Hope's replacement on the entry of every other creature
// under the controller's control, adding the Biomancer's power as it is at
// that moment, plus a Mutant subtype given as it enters (game.
// ReplacementEvent.AddSubtypeAtETB, the ADR 0210 entry overwrite) so no
// "enters" trigger or state check sees the creature without it. Tokens run
// the same entry pipeline and get both. A Biomancer with power 0 or less
// adds no counters but still adds the type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "afeba4d7-b659-4fb1-84dd-27704fbb2621",
		Name:         "Master Biomancer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			bnsB13OtherCreaturesEnterWithPowerCountersAsMutants(
				"Master Biomancer: enters with +1/+1 counters equal to its power and as a Mutant"),
		},
	})
}
