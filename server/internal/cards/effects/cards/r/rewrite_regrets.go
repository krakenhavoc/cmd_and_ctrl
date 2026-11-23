package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "42565432-3a17-4274-87fc-016ab900e313",
		Name:         "Rewrite Regrets",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target creature or planeswalker card with mana value 6 or less from your graveyard",
			effects.Or(effects.Creature(), effects.Planeswalker()), effects.YouOwn(), effects.ManaValueLE(6)),
		Purpose: effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return effects.EmpowerJace{N: 2}.Apply(ctx)
		},
	})
}
