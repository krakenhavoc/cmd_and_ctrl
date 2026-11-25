package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_auction.go — life bidding in turn order (#6212): Pain's Reward,
// Mages' Contest, Illicit Auction. The engine half is
// game/life_auction.go: the rounds, the pass, the end when the high bid
// stands. The card's tail is told the high bidder (o.IDs[0], uuid.Nil
// when they have left the game) and the high bid (o.Amount).

// BidLife starts the bidding.
//
//	// Pain's Reward
//	BidLife{OpenerChooses: true, Label: "Pain's Reward", Then: painsRewardWon}
//	// Illicit Auction
//	BidLife{OpeningBid: 0, Label: "Illicit Auction", Then: illicitAuctionWon}
//	// Mages' Contest: "You and target spell's controller bid life."
//	BidLife{Rival: spellController, OpeningBid: 1, Label: "Mages' Contest", Then: magesContestWon}
type BidLife struct {
	// Rival limits the bidders to you and this player ("You and target
	// spell's controller bid life"). Zero: each player still in the game
	// bids, in turn order starting with you.
	Rival uuid.UUID
	// OpenerChooses is "You start the bidding with a bid of any number";
	// otherwise you start with OpeningBid.
	OpenerChooses bool
	OpeningBid    int
	// Label names the card in each prompt.
	Label string
	// Then is the rest of the card, run against the resolving item (its
	// targets are read there): a CardTail.
	Then game.TailRef
}

// Apply starts the auction. Caller is inside the resolution frame.
func (b BidLife) Apply(ctx *Context) error {
	you := ctx.Controller()
	bidders := []uuid.UUID{you}
	if b.Rival == uuid.Nil {
		bidders = ctx.Game.TurnOrderFromForEffect(you)
	} else if b.Rival != you {
		bidders = append(bidders, b.Rival)
	}
	return ctx.Game.StartLifeAuctionForEffect(game.LifeAuction{
		Source:        ctx.Source(),
		Bidders:       bidders,
		OpenerChooses: b.OpenerChooses,
		OpeningBid:    b.OpeningBid,
		Label:         b.Label,
		Then:          ctx.ItemTail(b.Then),
	})
}
