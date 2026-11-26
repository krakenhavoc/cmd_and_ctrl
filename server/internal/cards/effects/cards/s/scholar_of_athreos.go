package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scholar of Athreos — Creature — Human Cleric {2}{W}, 1/4 (Theros, S82 OLD b391, #5561):
//
//	"{2}{B}: Each opponent loses 1 life. You gain life equal to the life
//	 lost this way."
//
// Ob Nixilis of the Black Oath's +2 as an activated ability at {2}{B}: the
// drain counts what each opponent actually lost, and the controller gains
// that total, so a table of three opponents gives three life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18c5e7a0-4397-48bb-b993-1c3e3b08be2c",
		Name:         "Scholar of Athreos",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: Each opponent loses 1 life. You gain life equal to the life lost this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B21DrainEachOpponentAndGainTheTotal(g, item, 1)
			},
		}},
	})
}
