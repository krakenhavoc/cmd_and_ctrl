package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Full Flowering — Sorcery {X}{X}{G}:
//
//	"Populate X times. (To populate, create a token that's a copy of a
//	 creature token you control. Do this X times.)"
//
// Each populate is the previous one's continuation rather than a loop
// of Apply calls: a populate over several different tokens asks which
// to copy, and the next one must see the copy the answer made (a copy
// of the new token is a legal pick). X is announced at cast and read
// from the item. With X = 0, or no creature token, nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb3328b5-a759-4c30-8b1f-202980bb6f6d",
		Name:         "Full Flowering",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return populateTimes(ctx, ctx.X())
		},
	})
}

// populateTimes is "populate n times", chained so each one resolves
// (prompt included) before the next starts.
func populateTimes(ctx *Context, n int) error {
	if n <= 0 {
		return nil
	}
	return Populate{
		Then: func(c *Context, created []uuid.UUID) error {
			if len(created) == 0 {
				// Nothing was copied (CR 701.36b): every later populate
				// would find the same board and copy nothing too.
				return nil
			}
			return populateTimes(c, n-1)
		},
	}.Apply(ctx)
}
