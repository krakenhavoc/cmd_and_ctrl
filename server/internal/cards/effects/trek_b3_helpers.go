package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trek_b3_helpers.go — shapes shared within Star Trek batch 3 (#4973).
// Append-only.

// condPermanentReturnedToYourHand is the intervening "if a permanent was
// returned to your hand from the battlefield this turn" (Miles O'Brien,
// Transporter Ops; CR 603.4), read off the item's controller's turn
// tally (Game.PermanentReturnedToHandThisTurn) as the trigger resolves.
type CondPermanentReturnedToYourHand struct{}

func (CondPermanentReturnedToYourHand) Holds(ctx *Context) bool {
	return ctx.Game.PermanentReturnedToHandThisTurn(ctx.Controller())
}

func (CondPermanentReturnedToYourHand) Condition() (game.FootprintStep, bool) {
	return CondStep(game.FootprintCondPermanentReturned, 0), true
}

// permanentReturnedToYourHand is the same "if" as the trigger fires, for
// a When.
func PermanentReturnedToYourHand(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.PermanentReturnedToHandThisTurn(source.Controller)
}
