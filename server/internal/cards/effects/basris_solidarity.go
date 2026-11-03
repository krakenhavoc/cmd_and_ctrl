package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basri's Solidarity — Sorcery {1}{W} (Jumpstart Selesnya deck request,
// #4893):
//
//	"Put a +1/+1 counter on each creature you control."
//
// The set is the creatures the caster controls as the spell resolves, read
// before the first counter goes on (Abandoned Air Temple's body).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68281e06-64e3-490b-8691-8f4169614ec0",
		Name:         "Basri's Solidarity",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b11PutCounterOnEachCreatureYouControl(ctx.Game, item)
		},
	})
}
