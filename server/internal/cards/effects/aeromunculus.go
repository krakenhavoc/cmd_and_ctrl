package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aeromunculus — Creature — Homunculus Mutant {1}{G}{U}, 2/3 (Ravnica
// Allegiance, #5521):
//
//	"Flying
//	 {2}{G}{U}: Adapt 1. (If this creature has no +1/+1 counters on it,
//	 put a +1/+1 counter on it.)"
//
// Flying is printed data (CR 702.9b). Adapt is adaptN (CR 701.46a): the
// counters go on only if it has no +1/+1 counter as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a0cd19c1-9125-4627-83ba-d42c489b4e7d",
		Name:            "Aeromunculus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}{U}: Adapt 1.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{G}{U}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return adaptN(g, item, 1) },
		}},
	})
}
