package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call for Blood — Instant — Arcane {4}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Target creature gets -X/-X until end of turn, where X is the
//	 sacrificed creature's power."
//
// X is the sacrificed creature's last-known power (the 2005-02-01
// ruling, CR 608.2h), read off the payment record (ADR 0113 §1). A
// negative power is X = 0 (CR 107.1b), never a pump. Targets are chosen
// before costs are paid (CR 601.2c before 601.2h), so sacrificing the
// target itself is legal and leaves the spell with no legal target: it
// does not resolve (the same ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "f2782c21-ed99-4821-8938-cc539e783ec7",
		Name:           "Call for Blood",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.SacrificedPower()
			return vsPumpTheTarget(ctx, -x, -x, fmt.Sprintf("Call for Blood — -%d/-%d", x, x))
		},
	})
}
