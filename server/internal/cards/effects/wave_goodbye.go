package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wave Goodbye — Sorcery {2}{U}{U} (EDHREC rank 1515):
//
//	"Return each creature without a +1/+1 counter on it to its
//	 owner's hand."
//
// The counters deck's one-sided Evacuation. One simultaneous bounce
// of every creature with no +1/+1 counter — every player's, the
// caster's own un-countered creatures included, as printed; tokens
// bounced this way cease to exist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06ef46c6-00ba-40e6-b866-d0095ab83749",
		Name:         "Wave Goodbye",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BounceAllMatching{Match: And(Creature(), b13WithoutPlusOneCounter())}.Apply(ctx)
		},
	})
}
