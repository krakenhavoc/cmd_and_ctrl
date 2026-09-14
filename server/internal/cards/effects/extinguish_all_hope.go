package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extinguish All Hope — Sorcery {4}{B}{B} (EDHREC rank 3833):
//
//	"Destroy all nonenchantment creatures."
//
// The enchantress deck's one-sided Wrath: every creature that is not
// also an enchantment, read post-layer, so a Theros god that is a
// creature survives and an animated Opalescence-style enchantment
// does too. One simultaneous destruction event, so every dies
// trigger sees the whole batch; indestructible is honoured on the
// mass path since #470.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7ca25f3-7a22-477b-8546-4c2597d5d1ff",
		Name:         "Extinguish All Hope",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), Not(Enchantment()))}.Apply(ctx)
		},
	})
}
