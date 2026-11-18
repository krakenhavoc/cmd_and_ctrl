package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contortionist Troupe — Creature — Human {X}{G}, 0/0 (Midnight Hunt,
// #5497):
//
//	"This creature enters with X +1/+1 counters on it.
//	 Coven — At the beginning of your end step, if you control three or
//	 more creatures with different powers, put a +1/+1 counter on target
//	 creature you control."
//
// Benevolent Hydra's X counters (XCounters). The end-step trigger needs
// coven as it goes on the stack, where its target is chosen, and again as
// it resolves (CR 603.4).
//
// No simplification.
func init() {
	const label = "Contortionist Troupe — Coven: put a +1/+1 counter on target creature you control"
	Register(Spec{
		OracleID:                   "be9bee08-ac04-42bc-93be-5a77482a7690",
		Name:                       "Contortionist Troupe",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			Targeting(InterveningIf(midB05CovenCondition(), On(game.EventBeginEndStep,
				AllOf(ByYou, midB05YouHaveCoven), label, PlusOneCounterOnTheTarget)), TargetCreature("target creature you control", YouControl())),
		},
	})
}
