package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skitter Eel — Creature — Fish Crab {3}{U}, 3/3 (Ravnica: Clue Edition):
//
//	"{2}{U}: Adapt 2. (If this creature has no +1/+1 counters on it, put two
//	 +1/+1 counters on it.)"
//
// Pteramander's adapt (adaptN).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75132177-0881-4158-9d65-73f5ba9d4402",
		Name:         "Skitter Eel",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Adapt 2.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{U}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return adaptN(g, item, 2) },
		}},
	})
}
