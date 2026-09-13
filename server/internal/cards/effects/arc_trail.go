package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arc Trail — Sorcery for {1}{R}:
//
//	"Arc Trail deals 2 damage to any target and 1 damage to any
//	other target."
//
// S20 sub-PR 5's positional multi-target card: two slots of "any
// target", distinct, and the ORDER matters — the first pick takes
// 2, the second 1. Each slot is checked separately at resolution
// (CR 608.2b): if the 2-damage target left, the 1 still lands.
func init() {
	Register(Spec{
		OracleID:     "f1c26b25-371e-4fbf-a43d-7fd59a364d3a",
		Name:         "Arc Trail",
		Completeness: CompletenessFull,
		Targets:      TargetAny().WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			amounts := []int{2, 1}
			for i, t := range item.Targets {
				if i >= len(amounts) || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amounts[i]}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
