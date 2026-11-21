package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benevolent Hydra — Creature — Hydra {X}{G}{G}, 1/1 (EDHREC rank
// 2955):
//
//	"This creature enters with X +1/+1 counters on it.
//	 If one or more +1/+1 counters would be put on another creature
//	 you control, that many plus one +1/+1 counters are put on it
//	 instead.
//	 {T}, Remove a +1/+1 counter from this creature: Put a +1/+1
//	 counter on another target creature you control."
//
// Hardened Scales on an X-sized body. The X counters are the printed
// CR 614.1c entry clause and ride the CR 614 pipeline as one
// (XCounters, #1002); the replacement is Conclave Mentor's with the
// Hydra itself excluded (b28OtherCreaturesYouControlGetAnExtraCounter
// — "another"), so a second Hydra's X counters get the bonus from the
// first, through the entry event, and a Hardened Scales stacks in
// CR 616 order.
//
// The tap ability moves a counter: its cost removes a +1/+1 counter
// from the Hydra at announce (RemoveCountersFromThis, #625), and the
// counter it puts on the target is an ordinary placement, so the
// Hydra's own replacement makes it two. The Hydra is a printed 1/1,
// so spending its last counter leaves a 1/1, not a 0/0.
//
// "Another target creature" is the clause's own ExcludeSource, so the
// exclusion is by instance: the Hydra cannot target itself, as printed,
// and a second Benevolent Hydra is a legal target, as printed (#3172).
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "01dbf1bc-ca62-4fb6-959c-ef7c0dc03bb0",
		Name:                       "Benevolent Hydra",
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Completeness:               effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B28OtherCreaturesYouControlGetAnExtraCounter("Benevolent Hydra: +1 +1/+1 counter"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on another target creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect:  effects.B36CounterOnChosenAnimal,
		}},
	})
}
