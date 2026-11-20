package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Belligerent Hatchling — Creature — Elemental {3}{R/W}, 6/6 (Eventide, #5590):
//
//	"First strike
//	 This creature enters with four -1/-1 counters on it.
//	 Whenever you cast a red spell, remove a -1/-1 counter from this creature.
//	 Whenever you cast a white spell, remove a -1/-1 counter from this
//	 creature."
//
// The four counters are an entry replacement (Moonshadow's shape), and each
// colour is its own trigger: a spell of both colours removes two. A
// creature with no -1/-1 counter left removes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7bd4e504-541e-4b2c-9f1d-3c67957aef2e",
		Name:            "Belligerent Hatchling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Replacements:    []game.ReplacementEffect{eveB1HatchlingCounters("Belligerent Hatchling")},
		Triggered: []game.TriggeredAbility{
			eveB1RemoveCounterOnCast("Belligerent Hatchling", "R"),
			eveB1RemoveCounterOnCast("Belligerent Hatchling", "W"),
		},
	})
}
