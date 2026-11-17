package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruins Recluse — Creature — Spider 1/1 {1}{G} (March of the Machine, S82):
//
//	"Reach, deathtouch
//	 {3}{G}: Put a +1/+1 counter on this creature."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d99efecd-1419-434b-9de6-5fbb04f57dda",
		Name:            "Ruins Recluse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1},
			Cost:    ManaCost("{3}{G}"),
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
