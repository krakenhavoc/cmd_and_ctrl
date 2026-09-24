package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Champion of Lambholt — Creature — Human Warrior {1}{G}{G}, 1/1:
//
//	"Creatures with power less than this creature's power can't block
//	 creatures you control.
//	 Whenever another creature you control enters, put a +1/+1 counter
//	 on this creature."
//
// The reference card for a BLOCKER-side pair rule (#750, ADR 0045
// addendum Decision 11):
//
//	CantBlockAttackers(PowerLessThanSource(), ControlledBySourceController(), …)
//
// It binds every creature its controller controls, the Champion
// included, and compares the blocker's power with the Champion's own —
// both read live when blockers are declared, after every layer and
// counter, so the threshold grows with each counter the second clause
// adds. A blocker with EQUAL power may block ("less than" is strict).
// Nothing is re-checked after the declaration (CR 509.1b): shrinking
// the Champion after blocks are declared frees no one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c549b0fd-1e08-4873-952e-a14dc45a0fd2",
		Name:         "Champion of Lambholt",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBlockAttackers(PowerLessThanSource(), ControlledBySourceController(),
				"creatures with power less than Champion of Lambholt's can't block creatures its controller controls"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl,
				"Champion of Lambholt — put a +1/+1 counter on it", putCounterOnSelf),
		},
	})
}
