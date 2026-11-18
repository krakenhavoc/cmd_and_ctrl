package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winged Portent — Instant {1}{U}{U}:
//
//	"Cleave {4}{G}{U} (You may cast this spell for its cleave cost. If
//	 you do, remove the words in square brackets.)
//	 Draw a card for each creature you control [with flying]."
//
// The count is taken as the spell resolves, through the layered view, so
// a creature granted flying counts and one that lost it does not (CR
// 702.148a, CR 608.2h). The brackets touch no target clause, hence
// Cleave's nil.
func init() {
	Register(Spec{
		OracleID:         "378567f8-7aef-42b6-853e-8591641052e8",
		Name:             "Winged Portent",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Cleave("{4}{G}{U}", nil)},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			match := And(Creature(), HasKeyword("flying"))
			if ctx.PaidAltCost("cleave") {
				match = Creature()
			}
			return DrawCards{N: countYouControl(ctx, match)}.Apply(ctx)
		},
	})
}
