package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Killing Glare — Instant {X}{B}:
//
//	"Destroy target creature with power X or less."
//
// ADR 0109 §9 (#1842): the power bound reads the announced X, which is
// chosen before the target (CR 601.2b–c) and re-checked as the spell
// resolves (CR 608.2b) — a creature pumped past X in response is an
// illegal target and the spell does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b47891f-130a-4d76-b262-53f5e3bb6b95",
		Name:         "Killing Glare",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with power X or less").WithPowerAtMostX(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return DestroyTarget{Target: t.ID}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
