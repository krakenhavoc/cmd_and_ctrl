package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Requiting Hex — {B} Instant:
//
//	"As an additional cost to cast this spell, you may blight 1. (You
//	 may put a -1/-1 counter on a creature you control.)
//	 Destroy target creature with mana value 2 or less. If this spell's
//	 additional cost was paid, you gain 2 life."
//
// OptionalBlight(1) (#1703). The life is gained whether or not the
// destruction happens: the second sentence is not an "if you do" on the
// first, and a spell whose only target has become illegal does not
// resolve at all (CR 608.2b), which the engine handles before this
// body runs. No simplification.
func init() {
	Register(Spec{
		OracleID:      "eda16b31-33db-4ad2-901a-80fbe6273733",
		Name:          "Requiting Hex",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalBlight(1)},
		Targets:       TargetCreature("target creature with mana value 2 or less", ManaValueLE(2)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.BlightPaid() {
				return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
