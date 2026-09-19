package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mass Calcify — Sorcery {5}{W}{W} (EDHREC rank 4204):
//
//	"Destroy all nonwhite creatures."
//
// A one-sided wrath for a mono-white deck, and the reason it is worth
// seven mana rather than four: your board survives and three
// opponents' boards usually do not.
//
// "Nonwhite" is read off the CURRENT colour of each creature, not the
// printed one — the predicate runs against the post-layer
// characteristic, so a creature an effect has made white is spared
// and a white creature something has turned black is not. That is CR
// 105.2a working as intended and it is the only rules subtlety on the
// card.
//
// The sweep goes through DestroyAllMatching so every death is
// simultaneous (CR 704.3 / the S23 wrath posture): a Blood Artist
// sees all of them at once, and an indestructible or regenerating
// creature is filtered out by the engine's own destructible narrowing
// rather than by anything here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ab3996f-aa3f-4041-8634-8e197d51f108",
		Name:         "Mass Calcify",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), Not(OfColor("W")))}.Apply(ctx)
		},
	})
}
