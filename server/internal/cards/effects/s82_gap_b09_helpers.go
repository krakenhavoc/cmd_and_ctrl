package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b09_helpers.go — shared pieces for the S82 GAP batch 9 cards
// (Innistrad Remastered creatures, #5948).

// condGapB09YouHaveNoCardsInHand is the intervening "if you have no cards
// in hand" (CR 603.4), read as the trigger fires and again as it
// resolves. A hand's size is public; its contents are not read.
type CondGapB09YouHaveNoCardsInHand struct{}

func (CondGapB09YouHaveNoCardsInHand) Holds(ctx *Context) bool {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	return p != nil && (p.Hand == nil || p.Hand.Size() == 0)
}

func (CondGapB09YouHaveNoCardsInHand) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
