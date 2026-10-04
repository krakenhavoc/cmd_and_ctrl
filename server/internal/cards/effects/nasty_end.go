package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nasty End — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw two cards. If the sacrificed creature was legendary, draw three
//	 cards instead."
//
// "Was legendary" is read off the sacrificed creature as it last existed
// on the battlefield (CR 608.2h), from the payment record (ADR 0113 §1):
// a creature made legendary by an effect counts, and so does a
// sacrificed legendary commander.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "f73da5d8-fd15-4315-ad3c-c86c28087285",
		Name:           "Nasty End",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 2
			if sacrificedHadSupertype(ctx, "Legendary") {
				n = 3
			}
			return DrawCards{N: n}.Apply(ctx)
		},
	})
}
