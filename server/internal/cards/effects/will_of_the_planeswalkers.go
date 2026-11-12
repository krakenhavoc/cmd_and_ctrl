package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// will_of_the_planeswalkers.go is the ability word of March of the
// Machine Commander's five Paths (#5648):
//
//	"Will of the Planeswalkers — Starting with you, each player votes for
//	 planeswalk or chaos. If planeswalk gets more votes, planeswalk. If
//	 chaos gets more votes or the vote is tied, chaos ensues."
//
// An ability word has no rules meaning of its own (CR 207.2c): the
// sentence is ADR 0146's vote (CR 701.38) followed by one of two
// Planechase instructions. "Planeswalk" is the spell's controller
// planeswalking (game.PlaneswalkForEffect), which only the planar
// controller may do (CR 701.31a): a Path cast on another player's turn
// planeswalks nobody. "Chaos ensues" is
// CR 311.7's second case, "a resolving spell or ability says that chaos
// ensues" (game.ChaosEnsuesForEffect): each face-up plane's chaos ability
// triggers, controlled by the planar controller, and goes on the stack
// once the spell has finished resolving.
//
// At a table that is not playing Planechase there is no planar deck and
// no plane: the vote still happens (it is a vote, so "whenever players
// finish voting" and extra votes work as usual), and neither result does
// anything.

// WillOfThePlaneswalkers is the vote and what follows it, one step of a
// spell's resolution. Card is the spell's name, for the ballot's header.
type WillOfThePlaneswalkers struct {
	Card string
}

func (w WillOfThePlaneswalkers) Apply(ctx *Context) error {
	return Vote{
		Question: w.Card + " — vote for planeswalk or chaos",
		Words:    []string{"planeswalk", "chaos"},
		Then:     willOfThePlaneswalkersVoted,
	}.Apply(ctx)
}

// willOfThePlaneswalkersVoted reads the tally. An on-disk key: never
// renamed, never reused.
var willOfThePlaneswalkersVoted = VoteResultThen("vote/will-of-the-planeswalkers", func(ctx *Context, r game.VoteResult) error {
	if !r.MoreVotes(0, 1) {
		// Chaos got more votes, or the vote is tied.
		ctx.Game.ChaosEnsuesForEffect(ctx.Controller(), ctx.Source())
		return nil
	}
	// CR 701.31a: "A player may planeswalk only during a Planechase game.
	// Only the planar controller may planeswalk."
	if ps := ctx.Game.Planechase; ps == nil || ps.Controller != ctx.Controller() {
		return nil
	}
	return ctx.Game.PlaneswalkForEffect(ctx.Controller())
})
