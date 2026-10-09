package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arc Trail — Sorcery for {1}{R}:
//
//	"Arc Trail deals 2 damage to any target and 1 damage to any
//	other target."
//
// Two instances of the word "target", so two clauses (CR 601.2c, ADR
// 0065 §1): slot 0 takes 2, slot 1 takes 1, and slot 1 is Distinct
// ("any other target"). Each slot is checked separately at resolution
// (CR 608.2b): if the 2-damage target left, the 1 still lands. Before
// #2689 it was one clause of two targets dealt 2 and 1 by position,
// which a per-clause purpose entry could not describe (ADR 0126's
// amendment of 2026-10-08); the moves now name each pick's slot.
func init() {
	Register(Spec{
		OracleID:     "f1c26b25-371e-4fbf-a43d-7fd59a364d3a",
		Name:         "Arc Trail",
		Completeness: CompletenessFull,
		Targets:      Clauses(TargetAny(), Distinct(TargetAny())),
		Purpose:      ForTargets(DamageToTarget(0, 2), DamageToTarget(1, 1)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			amounts := []int{2, 1}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for slot, n := range amounts {
					t, ok := ctx.ClauseTarget(slot)
					if !ok {
						continue
					}
					if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
