package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b79_helpers.go — shared pieces for the S82 GAP batch 79 cards
// (Ravnica Remastered instants, #5948).

// gapB79PutTargetOnTopOfItsOwnersLibrary is "Put target <permanent> on top
// of its owner's library" (Warrant, Totally Lost): the first still-legal
// target goes to the top of its owner's library, and a target that has gone
// does nothing (CR 608.2b).
func gapB79PutTargetOnTopOfItsOwnersLibrary(_ *game.StackItem, ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ctx.Game.TuckToLibraryForEffect(id, false)
}
