package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Logic Knot — Instant {X}{U}{U}:
//
//	"Delve. Counter target spell unless its controller pays {X}."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// Delve pays X's generic too: the budget folds the announced X in
// (ADR 0100 §1), so Logic Knot at X = 3 with three cards in the
// graveyard costs {U}{U}. No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2da7acb-d80c-414c-9f7d-753a5d6ccad9",
		Name:         "Logic Knot",
		Completeness: CompletenessFull,
		Delve:        true,
		XMatters:     true,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			cost := "{" + strconv.Itoa(ctx.X()) + "}"
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     cost,
				Question: "Logic Knot — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
