package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glistening Dawn — Sorcery {2}{G}{G} (March of the Machine):
//
//	"Incubate X twice, where X is the number of lands you control."
//
// Two separate incubates (Incubate.Times), each an Incubator token with
// X +1/+1 counters, where X is the number of lands the controller has as
// the spell resolves. X is read once: the first Incubator is not a land,
// so the count cannot change between the two, and the printed X is "the
// number of lands you control" at the one moment the instruction is
// carried out.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13ee85c3-a16a-41ac-bbaa-c7549c700a03",
		Name:         "Glistening Dawn",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			lands := 0
			for _, c := range ctx.Game.Battlefield.Cards {
				if c.Controller == ctx.Controller() && c.IsLand() {
					lands++
				}
			}
			return Incubate{N: lands, Times: 2}.Apply(ctx)
		},
	})
}
