package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patchwork Automaton — Artifact Creature — Construct {2}, 1/1 (Kamigawa:
// Neon Dynasty, S82):
//
//	"Ward {2}
//	 Whenever you cast an artifact spell, put a +1/+1 counter on this
//	 creature."
//
// Aboleth Spawn's ward {2} and Forensic Gadgeteer's artifact cast watcher,
// with a counter on the source (CounterOnThis, which does nothing once the
// Automaton has left).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ba86a50-13df-4b26-8b3a-9e3917ff850f",
		Name:         "Patchwork Automaton",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Patchwork Automaton — ward {2}"),
			WheneverYouCast(Artifact(), "Patchwork Automaton — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
