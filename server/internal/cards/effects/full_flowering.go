package effects

import (
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
// (prompt included) before the next starts. Keyed (#4253 slice 10b): the
// populates still owed are the tail's number.
func populateTimes(ctx *Context, n int) error {
	if n <= 0 {
		return nil
	}
	return Populate{ThenTail: withN(ctx.ItemTail(fullFloweringAgain), n)}.Apply(ctx)
}

// fullFloweringAgain is the next populate, told what this one created: none
// means every later populate would find the same board and copy nothing too
// (CR 701.36b). An on-disk key: never renamed, never reused. Registered from
// init: the body asks again under its own key.
var fullFloweringAgain game.TailRef

func init() {
	fullFloweringAgain = CardTail("then/full-flowering-populate-again", func(c *Context, o game.TailOutcome) error {
		if len(o.IDs) == 0 {
			return nil
		}
		return populateTimes(c, o.Tail.N-1)
	})
}
