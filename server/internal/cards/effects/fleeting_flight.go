package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6336401f-4e2d-4ebe-8c5b-24aa6f516abf",
		Name:         "Fleeting Flight",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		// Each sentence is a step: the keyword and the shield wait for a prompt
		// the counter asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Fleeting Flight — flying"}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return PreventDamageFromSource{Protect: ShieldObject(id), CombatOnly: true}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
