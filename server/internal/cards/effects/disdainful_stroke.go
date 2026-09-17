package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disdainful Stroke — Instant {1}{U} (EDHREC rank 2366):
//
//	"Counter target spell with mana value 4 or greater."
//
// The two-mana answer to the big spell. The clause is a stack target
// whose mana value is read ON THE STACK (CR 202.3e): ManaValueGE goes
// through game.(*Game).ManaValueForEffect, so an X spell counts what
// was announced for X — a Hydra cast for X=5 is a legal target, one
// cast for X=1 is not. The engine re-checks the same predicate at
// resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11e02134-7b1a-46a4-a89e-7539dd1efada",
		Name:         "Disdainful Stroke",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell with mana value 4 or greater", ManaValueGE(4)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
