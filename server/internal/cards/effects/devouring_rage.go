package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devouring Rage — Instant — Arcane {4}{R}:
//
//	"As an additional cost to cast this spell, you may sacrifice any
//	 number of Spirits. Target creature gets +3/+0 until end of turn.
//	 For each Spirit sacrificed this way, that creature gets an
//	 additional +3/+0 until end of turn."
//
// The variable sacrifice (ADR 0100 §3): "you may sacrifice any number"
// is one clause whose count runs from zero, because sacrificing none is
// not paying. The Spirits die with the spell on the stack; the count is
// read back from the announcement (ctx.Sacrificed). The two sentences
// are one boost of 3 + 3 × the count, which is the same +N/+0 until end
// of turn in layer 7c.
//
// Arcane matters only to splice, which reads the type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "19a8c2d3-b06e-4a76-bb28-8928ce84186c",
		Name:           "Devouring Rage",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeAnyNumberCost("any number of Spirits", HasSubtype("Spirit")),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 3 + 3*ctx.Sacrificed()
			return vsPumpTheTarget(ctx, n, 0, fmt.Sprintf("Devouring Rage — +%d/+0", n))
		},
	})
}
