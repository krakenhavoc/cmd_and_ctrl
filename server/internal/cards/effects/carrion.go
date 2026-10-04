package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carrion — Instant {1}{B}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Create X 0/1 black Insect creature tokens, where X is the sacrificed
//	 creature's power."
//
// X is the sacrificed creature's power as it last existed on the
// battlefield (CR 608.2h), read off the payment record (ADR 0113 §1); a
// power of zero or less makes none (CR 107.1b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "84ca8d74-95ec-4a9c-8793-7a195ad28be9",
		Name:           "Carrion",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("0/1 black Insect"), N: ctx.SacrificedPower()}.Apply(ctx)
		},
	})
}
