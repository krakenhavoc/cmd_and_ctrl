package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "25781dc5-d6ea-4273-9fe3-06a8808b3718",
		Name:         "Reject Imperfection",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			small := ctx.Game.StackItemForEffect(stackID) != nil && effects.S82OneB08ManaValueAtMost(ctx, stackID, 3)
			if err := (effects.CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			if !small {
				return nil
			}
			return effects.Proliferate{}.Apply(ctx)
		},
	})
}
