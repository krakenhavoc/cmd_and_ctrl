package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horseshoe Crab — Creature — Crab {2}{U}, 1/3 (Dominaria Remastered,
// #5948):
//
//	"{U}: Untap this creature."
//
// An ordinary activated ability (not a mana ability), so it uses the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "793f29ac-8a5d-49bf-91e8-a32f771268f3",
		Name:         "Horseshoe Crab",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			UntapSelf: true,
			Label:     "{U}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
