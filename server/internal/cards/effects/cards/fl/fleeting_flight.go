package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleeting Flight — Instant {W}:
//
//	"Put a +1/+1 counter on target creature. It gains flying until end of
//	 turn. Prevent all combat damage that would be dealt to it this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the counter and the keyword, then
// a combat-only not-one-use shield pinned to the creature, so its
// non-combat damage is still dealt.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6336401f-4e2d-4ebe-8c5b-24aa6f516abf",
		Name:         "Fleeting Flight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		// Each sentence is a step: the keyword and the shield wait for a prompt
		// the counter asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
			effects.Instruction(func(ctx *effects.Context) error {
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Fleeting Flight — flying"}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					return effects.PreventDamageFromSource{Protect: effects.ShieldObject(id), CombatOnly: true}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
