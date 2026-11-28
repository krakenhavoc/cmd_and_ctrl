package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vodalian Mage — Creature — Merfolk Wizard {2}{U}, 1/1 (Fallen Empires, S82 OLD
// b856, #5637):
//
//	"{U}, {T}: Counter target spell unless its controller pays {1}."
//
// Miscalculation's soft counter at {1}, on a creature's activated ability: the
// spell's controller is asked as it resolves, and a "pay" they can't fund is a
// decline. A spell that has already left the stack is a fizzle (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8dee4577-0315-4c9f-ab1c-0c68e2ab6a76",
		Name:         "Vodalian Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Counter target spell unless its controller pays {1}.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.TargetSpell("target spell"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     "{1}",
					Question: "Vodalian Mage — pay {1} or your spell is countered",
				}.Apply(ctx)
			},
		}},
	})
}
