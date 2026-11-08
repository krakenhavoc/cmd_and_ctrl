package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hazard of the Dunes — Creature — Wurm {3}{G}, 4/4 (Aetherdrift, #4464):
//
//	"Reach, trample
//	 Exhaust — {6}{G}: Put three +1/+1 counters on this creature.
//	 (Activate each exhaust ability only once.)"
//
// Greenbelt Guardian's exhaust at {6}{G} (CR 702.177a: "Activate only
// once"), keyed to the object, so a Wurm back from a flicker may use it
// again (CR 400.7). Reach and trample are the engine's (CR 702.17,
// CR 702.19).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "799664fc-279c-40c1-8e40-44a68c38801b",
		Name:            "Hazard of the Dunes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {6}{G}: Put three +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 3},
			Exhaust: true,
			Cost:    ManaCost("{6}{G}"),
			Effect:  plusOneCountersOnThis(3),
		}},
	})
}
