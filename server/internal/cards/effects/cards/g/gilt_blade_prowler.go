package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gilt-Blade Prowler — Creature — Human Rogue {2}{B}, 2/3 (Modern Horizons
// 2, #5501):
//
//	"{1}, {T}, Pay 1 life: Draw a card. Activate only if you've discarded a
//	 card this turn."
//
// The restriction reads the activator's discard tally for the turn (one per
// card, CR 701.9a) when the ability is announced; the life is a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63e86054-5e9c-48fe-92a1-b71aeb6ee57f",
		Name:         "Gilt-Blade Prowler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Pay 1 life: Draw a card. Activate only if you've discarded a card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.PayLife(1)),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.OldB030DiscardedACardThisTurn(g, controller)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
