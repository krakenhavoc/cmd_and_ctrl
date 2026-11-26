package cardsso

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spell Rupture — Instant {1}{U} (Gatecrash, #5565):
//
//	"Counter target spell unless its controller pays {X}, where X is the
//	 greatest power among creatures you control."
//
// "You" is Spell Rupture's controller, and X is read as it resolves
// (Thassa's Rebuff's shape): the greatest layered power among the creatures
// they control, zero with none, and with X = 0 the spell's controller pays
// nothing and the spell is not countered. Being a pay-unless it is legal
// against a spell that can't be countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "934c50aa-7d2e-4c9e-b91b-1e8c4055e1c6",
		Name:         "Spell Rupture",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			x := effects.B42GreatestPowerControlledBy(ctx.Game, ctx.Controller())
			if x <= 0 {
				return nil
			}
			cost := "{" + strconv.Itoa(x) + "}"
			return effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     cost,
				Question: "Spell Rupture — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
