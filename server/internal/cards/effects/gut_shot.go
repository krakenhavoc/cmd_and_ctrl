package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gut Shot — Instant {R/P} (EDHREC rank 2260):
//
//	"({R/P} can be paid with either {R} or 2 life.)
//	 Gut Shot deals 1 damage to any target."
//
// The free ping. One damage to any target from the spell as its
// source — Lightning Bolt's shape at one.
//
// Sandbox simplification, WEAKER than printed, the Gitaxian Probe
// posture: the Phyrexian symbol is charged as {R}. ParseCost records
// {R/P} as a red requirement flagged Phyrexian, but no payment path
// consults the flag, so the "or 2 life" option does not exist yet —
// the spell costs one red mana and cannot be cast for free. A
// mana-payment seam, not a card file's.
func init() {
	Register(Spec{
		OracleID:     "9afb3b6e-4909-4efa-aa79-81c0229411c9",
		Name:         "Gut Shot",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Phyrexian mana isn't supported — you must pay {R}, you can't pay 2 life instead."},
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 1}.Apply(ctx)
		},
	})
}
