package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Planar Engineering — Sorcery {3}{G} (EDHREC rank 3542):
//
//	"Sacrifice two lands. Search your library for four basic land
//	 cards, put them onto the battlefield tapped, then shuffle."
//
// Two lands into four tapped basics, and a landfall deck's best
// friend. The sacrifice is the caster's own choice — two sacrifice
// prompts over their lands, one land each (the b17PlayerSacrificesN
// shape); a caster with one land sacrifices it and the second prompt
// is skipped, and the search still happens, as CR 608.2c reads —
// and the search is a real prompt over the library's basic land
// cards, up to four, put onto the battlefield tapped, then a
// shuffle.
//
// Shape note, not a caveat: the search prompt is queued alongside
// the sacrifice prompts rather than after them, because the
// sacrifice prompt has no continuation. The lands leave the
// battlefield and the basics come from the library, so no answer to
// one changes the options of the other.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc48079a-b1c4-4c43-a989-418c78a264ec",
		Name:         "Planar Engineering",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b33SacrificeLandsThenSearchBasicsTapped(item, ctx, 2, 4)
		},
	})
}
