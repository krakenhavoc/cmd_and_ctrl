package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reject Imperfection — Instant {1}{U}{U} (Phyrexia: All Will Be One, #5479):
//
//	"Counter target spell. If that spell's mana value was 3 or less,
//	 proliferate."
//
// The spell's mana value is read while it is still on the stack (CR 202.3e,
// so an X spell counts its X) before it is countered, as last known
// information (CR 608.2h). The proliferate happens whether or not the
// counter took effect (an uncounterable spell still has its mana value).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25781dc5-d6ea-4273-9fe3-06a8808b3718",
		Name:         "Reject Imperfection",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			small := ctx.Game.StackItemForEffect(stackID) != nil && s82OneB08ManaValueAtMost(ctx, stackID, 3)
			if err := (CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			if !small {
				return nil
			}
			return Proliferate{}.Apply(ctx)
		},
	})
}
