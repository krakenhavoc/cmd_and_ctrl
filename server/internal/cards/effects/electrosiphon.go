package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Electrosiphon — Instant {U}{U}{R}:
//
//	"Counter target spell. You get an amount of {E} (energy counters)
//	 equal to its mana value."
//
// ADR 0129 PR 1. The mana value is read before the spell is countered,
// off the spell on the stack, X included (CR 202.3e). A spell that can't
// be countered still gives the energy: the two sentences are separate
// instructions. If the target is illegal the whole spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bff9105f-d8d4-4412-88d9-c9697581ebe9",
		Name:         "Electrosiphon",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			target := item.Targets[0].ID
			mv := 0
			if spell, ok := ctx.Game.LookupCardForEffect(target); ok {
				mv, _ = ctx.Game.ManaValueForEffect(spell)
			}
			if err := (CounterTarget{StackID: target}).Apply(ctx); err != nil {
				return err
			}
			return GetEnergy{N: mv}.Apply(ctx)
		},
	})
}
