package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skystreak Engineer — Creature — Human Pilot {1}{U}, 1/3 (Aetherdrift,
// #4464):
//
//	"Flying
//	 Exhaust — {4}{U}: Put two +1/+1 counters on this creature.
//	 (Activate each exhaust ability only once.)"
//
// Prowcatcher Specialist's exhaust row in blue, at a higher price: the
// engine's exhaust bit refuses a second activation by the same object,
// and the counters stay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2f920bf2-6a49-47c2-be3c-9787c377673b",
		Name:            "Skystreak Engineer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {4}{U}: Put two +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 2},
			Exhaust: true,
			Cost:    ManaCost("{4}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
