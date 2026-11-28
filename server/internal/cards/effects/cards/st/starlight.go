package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starlight — Sorcery {1}{W} (Portal, S82 OLD batch 784, #5630):
//
//	"You gain 3 life for each black creature target opponent controls."
//
// Renewing Dawn's count over black creatures: counted once as the spell
// resolves, read through the layered view, and the caster gains three for each.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6237896e-f033-41b2-9943-670f1becb582",
		Name:         "Starlight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB784GainLifePerTargetOpponentsPermanent(ctx, 3, effects.And(effects.Creature(), effects.OfColor("B")))
		},
	})
}
