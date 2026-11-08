package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Road Rage — Instant {R} (Aetherdrift, #4464):
//
//	"Road Rage deals X damage to target creature or planeswalker, where X
//	 is 2 plus the number of Mounts and Vehicles you control."
//
// X is counted once, as the spell resolves (CR 608.2h), over the
// permanents its controller controls then, by effective subtype: a
// Vehicle counts whether or not it is crewed, and a permanent that is
// both a Mount and a Vehicle counts once. The damage is
// DealDamageToTheTarget's, so a target that left in response is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da052082-b440-4abb-9c69-05d4d13e5fd2",
		Name:         "Road Rage",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// "Mounts" is a word of the spell's text (CR 612.2); Vehicle
			// is an artifact type, which no text change reaches.
			mount := ctx.TextSubtype("Mount")
			x := 2 + countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool {
				return c.HasSubtype(mount) || c.HasSubtype("Vehicle")
			})
			return DealDamageToTheTarget(x)(ctx.Game, item)
		},
	})
}
