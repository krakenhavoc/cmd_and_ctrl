package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ionize — Instant {1}{U}{R}:
//
//	"Counter target spell. Ionize deals 2 damage to that spell's
//	 controller."
//
// The spell's controller is read before the counter removes it from the
// stack. If the spell has already left the stack the target is illegal
// and Ionize does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01da67a5-e3dc-44f4-8a93-513bfa4d8fbc",
		Name:         "Ionize",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			spell := ctx.Game.StackItemForEffect(stackID)
			if spell == nil {
				return nil
			}
			controller := spell.Controller
			if err := (CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			return DealDamage{Source: ctx.Source(), Target: controller, Amount: 2}.Apply(ctx)
		},
	})
}
