package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rewrite Regrets — Sorcery {3}{B} (Reality Fracture):
//
//	"Return target creature or planeswalker card with mana value 6 or
//	 less from your graveyard to the battlefield.
//	 Empower Jace 2."
//
// ADR 0139 proof card. "To the battlefield" with no controller clause
// is under its owner's control, and the target is in YOUR graveyard,
// so that is you.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42565432-3a17-4274-87fc-016ab900e313",
		Name:         "Rewrite Regrets",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target creature or planeswalker card with mana value 6 or less from your graveyard",
			Or(Creature(), Planeswalker()), YouOwn(), ManaValueLE(6)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			reanimateSingleTarget(ctx, ctx.Controller())
			return EmpowerJace{N: 2}.Apply(ctx)
		},
	})
}
