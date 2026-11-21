package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirror of Galadriel — Legendary Artifact {2} (The Lord of the Rings:
// Tales of Middle-earth):
//
//	"{5}, {T}: Scry 1, then draw a card. This ability costs {1} less to
//	 activate for each legendary creature you control."
//
// Lembas's scry-then-draw (the draw is the scry's tail, so it waits for the
// reorder). The discount is a cost modifier on the ability itself, counting
// the activator's legendary creatures live as the ability is priced
// (CR 602.2b, 601.2f): the generic part only, never below {0}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7695ef41-fa58-4890-b3fe-de5751831d8e",
		Name:         "Mirror of Galadriel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Scry 1, then draw a card. This ability costs {1} less to activate for each legendary creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.PermanentsYouControl(effects.And(effects.Creature(), effects.Legendary())),
					"This ability costs {1} less to activate for each legendary creature you control."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.Scry{Player: ctx.Controller(), N: 1, ThenTail: ctx.TailN(effects.ThenDrawCards, 1)}.Apply(ctx)
			},
		}},
	})
}
