package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Even the Odds — Instant {2}{W} (Future Sight, #5594):
//
//	"Cast this spell only if you control fewer creatures than each
//	 opponent.
//	 Create three 1/1 white Soldier creature tokens."
//
// The cast restriction is a CastCondition (CR 601.3), read as the spell is
// announced: the caster must control fewer creatures than every opponent
// still in the game, compared one opponent at a time. The tokens are
// Captain's Call's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "6cd3605c-7885-4d32-8585-3b0b0f733702",
		Name:               "Even the Odds",
		Completeness:       effects.CompletenessFull,
		Purpose:            game.Purpose{Tokens: 3},
		CastCondition:      effects.FutB1YouControlFewerCreaturesThanEachOpponent,
		CastConditionLabel: "Cast this spell only if you control fewer creatures than each opponent.",
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 white Soldier"), N: 3}.Apply(ctx)
		},
	})
}
