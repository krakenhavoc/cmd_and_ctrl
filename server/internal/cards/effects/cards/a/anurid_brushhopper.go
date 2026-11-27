package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anurid Brushhopper — Creature — Frog Beast {1}{G}{W}, 3/4 (Judgment, #5611):
//
//	"Discard two cards: Exile this creature. Return it to the battlefield under
//	 its owner's control at the beginning of the next end step."
//
// Fleeting Spirit's blink with a two-card discard for its cost.
//
// No simplification.
func init() {
	const name = "Anurid Brushhopper"
	effects.Register(effects.Spec{
		OracleID:     "bdecf59c-f8bd-4003-9058-685b70c08837",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard two cards: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.DiscardN(2, "two cards"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileThisThenScheduleReturn(effects.NewContext(g, item), item, name+" — return Anurid Brushhopper")
			},
		}},
	})
}
