package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Channeler Initiate — Creature — Human Druid {1}{G}, 3/4 (Lorwyn Eclipsed
// Commander, #5368):
//
//	"When this creature enters, put three -1/-1 counters on target
//	 creature you control.
//	 {T}, Remove a -1/-1 counter from this creature: Add one mana of any
//	 color."
//
// The enters trigger has a required target, a creature you control; it may
// be the Initiate itself, which is then a 0/1 that can pay for its mana
// three times. The mana ability costs {T} and a -1/-1 counter removed at
// announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a98ea782-1c49-4207-a130-41329087fdf6",
		Name:         "Channeler Initiate",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Channeler Initiate — put three -1/-1 counters on target creature you control", cmdr2B25CountersOnTarget(3)),
				TargetCreature("target creature you control", YouControl()),
			),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, RemoveCounters: RemoveCountersFromThis(game.CounterMinusOne, 1).RemoveCounters},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Remove a -1/-1 counter from this creature: Add one mana of any color.",
		}},
	})
}
