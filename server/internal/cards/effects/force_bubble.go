package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Force Bubble — Enchantment {2}{W}{W}:
//
//	"If damage would be dealt to you, put that many depletion counters
//	 on this enchantment instead.
//	 When there are four or more depletion counters on this enchantment,
//	 sacrifice it.
//	 At the beginning of each end step, remove all depletion counters
//	 from this enchantment."
//
// ADR 0107 §1 (#1858). The first line is a CR 614.1a replacement ("instead"),
// not a prevention effect: it does not say "prevent" (CR 615.1a). Damage
// to the controller, combat or not, becomes that many counters. The
// sacrifice is a CR 603.8 state trigger: a single 5-damage hit puts five
// counters on and triggers it once, and the Bubble still absorbs whatever
// arrives before the trigger resolves. Every end step — anyone's —
// empties it again.
//
// No simplification.
func init() {
	const depletion = "depletion"
	Register(Spec{
		OracleID:     "655ae8e7-372b-4d8d-b33f-4aca46831abb",
		Name:         "Force Bubble",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			damageToYouBecomesCounters(depletion, true, "Force Bubble — put that many depletion counters on it instead"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisHasAtLeast(depletion, 4, "Force Bubble — sacrifice it", SacrificeThisIfStillOnBattlefield),
			AtEachStep(game.StepEnd, "Force Bubble — remove all depletion counters", removeAllCountersFromThis(depletion)),
		},
	})
}
