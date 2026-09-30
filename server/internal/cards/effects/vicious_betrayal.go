package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vicious Betrayal — Sorcery {3}{B}{B}:
//
//	"As an additional cost to cast this spell, sacrifice any number of
//	 creatures. Target creature gets +2/+2 until end of turn for each
//	 creature sacrificed this way."
//
// The variable sacrifice (ADR 0100 §3): the caster names any number of
// their creatures, zero included, and they die while the spell is on
// the stack, so a Blood Artist drains before the pump lands. The count
// is read back from the announcement (ctx.Sacrificed), because by
// resolution the creatures are in graveyards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "be1015a7-2ace-4b97-8884-202abae8401b",
		Name:           "Vicious Betrayal",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeAnyNumberCost("any number of creatures", Creature()),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 2 * ctx.Sacrificed()
			return vsPumpTheTarget(ctx, n, n, fmt.Sprintf("Vicious Betrayal — +%d/+%d", n, n))
		},
	})
}
