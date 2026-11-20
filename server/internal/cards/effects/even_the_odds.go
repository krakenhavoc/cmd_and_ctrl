package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:           "6cd3605c-7885-4d32-8585-3b0b0f733702",
		Name:               "Even the Odds",
		Completeness:       CompletenessFull,
		Purpose:            game.Purpose{Tokens: 3},
		CastCondition:      futB1YouControlFewerCreaturesThanEachOpponent,
		CastConditionLabel: "Cast this spell only if you control fewer creatures than each opponent.",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 white Soldier"), N: 3}.Apply(ctx)
		},
	})
}
