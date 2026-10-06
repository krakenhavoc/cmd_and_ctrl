package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lithomancer's Focus — Instant {W}:
//
//	"Target creature gets +2/+2 until end of turn. Prevent all damage
//	 that would be dealt to that creature this turn by colorless
//	 sources."
//
// The shield protects the target and reads #2026's colourless test (CR
// 105.2c) as each source would deal damage (CR 609.7b; the ruling: it
// prevents damage "from any colorless source that turn, even if that
// source didn't exist or wasn't colorless as Lithomancer's Focus
// resolved"). Combat damage or not.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a8a27a87-fc4c-44b5-9ba8-505bbf42164e",
		Name:         "Lithomancer's Focus",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := boostTheTargetUntilEOT(2, 2, "Lithomancer's Focus — +2/+2")(item, ctx); err != nil {
				return err
			}
			return PreventDamageFromSource{Protect: ShieldTheTarget, Filter: game.DamageSourceFilter{Colorless: true}}.Apply(ctx)
		},
	})
}
