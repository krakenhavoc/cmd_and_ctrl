package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// We Say Thee Nay! — {1}{U} Instant — Arcane:
//
//	"Teamwork 2 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 2 or more.)
//	 Counter target spell unless its controller pays {2}. Counter that
//	 spell unless its controller pays {4} instead if this spell was
//	 cast using teamwork."
//
// #1703: Teamwork(2). "Unless its controller pays" is CounterUnlessPaid
// (#951, Daze / Spell Stutter's shape): the prompt goes to the targeted
// spell's controller, not this spell's caster, a "pay" they can't fund
// degrades to a decline server-side, and the prompt holds the stack
// while it is unanswered. The tax is read once, at resolution, off
// whether teamwork was announced with this cast. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1abe8246-d2f9-407b-8c16-83da0a6b7de3",
		Name:          "We Say Thee Nay!",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Teamwork(2)},
		Targets:       effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			cost := "{2}"
			if ctx.UsedTeamwork() {
				cost = "{4}"
			}
			return effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     cost,
				Question: "We Say Thee Nay! — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
