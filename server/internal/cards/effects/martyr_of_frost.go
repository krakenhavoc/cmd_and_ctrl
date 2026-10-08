package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Martyr of Frost — Creature — Human Wizard {U}, 1/1:
//
//	"{2}, Reveal X blue cards from your hand, Sacrifice this creature:
//	 Counter target spell unless its controller pays {X}."
//
// The reveal cost is effects.RevealX (#2598, ADR 0020's 2026-10-08
// amendment): its count is the X announced with the activation, the
// revealed cards stay in the hand, and the mana cost is {2} whatever X
// is. The tax is the same X (CounterUnlessPaid, as Logic Knot). At
// X = 0 the spell's controller pays {0}, which is always possible, so
// nothing is asked and nothing is countered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "196cacb0-a26a-4540-9143-57867a638085",
		Name:         "Martyr of Frost",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Reveal X blue cards from your hand, Sacrifice this creature: Counter target spell unless its controller pays {X}.",
			Cost:    Plus(ManaCost("{2}"), RevealX("X blue cards", "U"), SacrificeThis()),
			Targets: TargetSpell("target spell"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				if x <= 0 || len(item.Targets) == 0 {
					return nil
				}
				cost := "{" + strconv.Itoa(x) + "}"
				return CounterUnlessPaid{
					StackID:  item.Targets[0].ID,
					Cost:     cost,
					Question: "Martyr of Frost — pay " + cost + " or your spell is countered",
				}.Apply(ctx)
			},
		}},
	})
}
