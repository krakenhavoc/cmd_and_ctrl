package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sauroform Hybrid — Creature — Human Lizard Warrior {1}{G}, 2/2 (Ravnica:
// Clue Edition):
//
//	"{4}{G}{G}: Adapt 4. (If this creature has no +1/+1 counters on it, put
//	 four +1/+1 counters on it.)"
//
// Pteramander's adapt (adaptN), which checks for counters as it resolves
// (CR 701.46a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "559c9949-faaf-41bc-9349-b6738b4763a2",
		Name:         "Sauroform Hybrid",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}{G}: Adapt 4.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{4}{G}{G}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return adaptN(g, item, 4) },
		}},
	})
}
