package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Homarid — Creature {2}{U}, 2/2:
//
//	"This creature enters with a tide counter on it.
//	 At the beginning of your upkeep, put a tide counter on this
//	 creature.
//	 As long as there is exactly one tide counter on this creature, it
//	 gets -1/-1.
//	 As long as there are exactly three tide counters on this creature,
//	 it gets +1/+1.
//	 Whenever there are four or more tide counters on this creature,
//	 remove all tide counters from it."
//
// ADR 0107 §1 (#1858). A four-turn tide cycle: 1/1, 2/2, 3/3, then the
// fourth counter resets it. The entry counter is a CR 614.1c
// replacement, the two statics are layer-7c effects read off the count
// as it stands, and the reset is a CR 603.8 state trigger — "whenever"
// in the printed text, but a state, so it triggers once when the fourth
// counter lands and not again while it waits. Between the fourth counter
// and the reset the creature is a 2/2: no static applies at four.
//
// No simplification.
func init() {
	const tide = "tide"
	Register(Spec{
		OracleID:     "5d02b0d9-cc53-4967-84bb-8b7f5c489b5c",
		Name:         "Homarid",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(tide, 1, "Homarid: enters with a tide counter"),
		},
		Static: []game.StaticAbility{
			whileExactlyCounters(tide, 1, thisCreatureOnly, -1, -1),
			whileExactlyCounters(tide, 3, thisCreatureOnly, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Homarid — put a tide counter", putACounterOnThis(tide)),
			WhenThisHasAtLeast(tide, 4, "Homarid — remove all tide counters", removeAllCountersFromThis(tide)),
		},
	})
}
