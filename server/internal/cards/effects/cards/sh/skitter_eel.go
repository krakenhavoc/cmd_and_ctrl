package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skitter Eel — Creature — Fish Crab {3}{U}, 3/3 (Ravnica: Clue Edition):
//
//	"{2}{U}: Adapt 2. (If this creature has no +1/+1 counters on it, put two
//	 +1/+1 counters on it.)"
//
// Pteramander's adapt (adaptN).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75132177-0881-4158-9d65-73f5ba9d4402",
		Name:         "Skitter Eel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Adapt 2.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 2) },
		}},
	})
}
