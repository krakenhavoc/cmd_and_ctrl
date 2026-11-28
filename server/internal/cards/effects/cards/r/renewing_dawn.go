package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renewing Dawn — Sorcery {1}{W} (Portal, S82 OLD batch 784, #5630):
//
//	"You gain 2 life for each Mountain target opponent controls."
//
// The target opponent's Mountains are counted once as the spell resolves, by
// effective land subtype; the caster gains that much. A target that is illegal
// on resolution makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54ea46ea-7c83-44a9-85b0-eff9745c6ffa",
		Name:         "Renewing Dawn",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB784GainLifePerTargetOpponentsPermanent(ctx, 2, effects.Subtype("Mountain"))
		},
	})
}
