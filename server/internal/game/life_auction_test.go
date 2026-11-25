package game

import (
	"testing"

	"github.com/google/uuid"
)

// life_auction_test.go — #6212: life bidding in turn order (Pain's
// Reward, Mages' Contest, Illicit Auction). The card's text is the
// procedure; CR 101.4 is the turn order the bidders are asked in.

var (
	auctionWinner uuid.UUID
	auctionBid    = -1
	auctionTold   = RegisterTail("test/life-auction-told", func(_ *Game, o TailOutcome) error {
		auctionWinner, auctionBid = uuid.Nil, o.Amount
		if len(o.IDs) > 0 {
			auctionWinner = o.IDs[0]
		}
		return nil
	})
)

func resetAuction() { auctionWinner, auctionBid = uuid.Nil, -1 }

// bid answers the open pay_amount prompt, which must be `who`'s.
func bid(t *testing.T, g *Game, who *Player, amount int) {
	t.Helper()
	c := openPromptOfKind(g, PendingChoicePayAmount)
	if c == nil {
		t.Fatalf("no bid prompt is open for %s", who.Name)
	}
	if c.Chooser != who.ID {
		t.Fatalf("the bid prompt is %v's, want %s's", c.Chooser, who.Name)
	}
	if err := g.ResolvePayAmount(c.ID, who.ID, amount); err != nil {
		t.Fatalf("%s bids %d: %v", who.Name, amount, err)
	}
}

func startAuction(t *testing.T, g *Game, a LifeAuction) {
	t.Helper()
	resetAuction()
	a.Then = KeyedTail{Ref: auctionTold}
	g.WithWriteLock(func() {
		if err := g.StartLifeAuctionForEffect(a); err != nil {
			t.Fatalf("StartLifeAuctionForEffect: %v", err)
		}
	})
}

// The issue's acceptance: Illicit Auction opens at 0; bids 0 → 3 → 5 and
// two passes leave the 5 standing, and the 5-bidder wins it.
func TestIllicitAuctionBidsInTurnOrderUntilTheHighBidStands(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	var bidders []uuid.UUID
	g.WithWriteLock(func() { bidders = g.TurnOrderFromForEffect(a.ID) })
	startAuction(t, g, LifeAuction{Bidders: bidders, OpeningBid: 0, Label: "Illicit Auction"})
	bid(t, g, b, 3) // tops 0
	bid(t, g, c, 5) // tops 3
	if c := openPromptOfKind(g, PendingChoicePayAmount); c == nil || c.PayAmount.Min != 6 {
		t.Fatalf("the next prompt = %+v, want a bid of 6 or more", c)
	}
	bid(t, g, a, 0) // pass
	if auctionBid >= 0 {
		t.Fatal("the bidding ended after one pass; the 3-bidder has not passed yet")
	}
	bid(t, g, b, 0) // pass: 5 stands
	if auctionWinner != c.ID || auctionBid != 5 {
		t.Fatalf("winner %v at %d, want %s at 5", auctionWinner, auctionBid, c.Name)
	}
	if openPromptOfKind(g, PendingChoicePayAmount) != nil {
		t.Fatal("a bid prompt is still open after the high bid stood")
	}
}

// Nobody tops the opening bid: it stands, and the opener wins.
func TestTheOpeningBidStandsWhenEveryonePasses(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	startAuction(t, g, LifeAuction{Bidders: []uuid.UUID{a.ID, b.ID, c.ID}, OpeningBid: 1, Label: "Mages' Contest"})
	bid(t, g, b, 0)
	bid(t, g, c, 0)
	if auctionWinner != a.ID || auctionBid != 1 {
		t.Fatalf("winner %v at %d, want the opener at 1", auctionWinner, auctionBid)
	}
}

// Pain's Reward: the opener chooses the opening bid; a bid must top the
// high bid, and the high bidder is not asked to top their own bid.
func TestPainsRewardOpenerChoosesAndTheHighBidderIsSkipped(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	a, b := g.Seats[0], g.Seats[1]
	startAuction(t, g, LifeAuction{Bidders: []uuid.UUID{a.ID, b.ID}, OpenerChooses: true, Label: "Pain's Reward"})
	open := openPromptOfKind(g, PendingChoicePayAmount)
	if open == nil || open.Chooser != a.ID || open.PayAmount.Min != 0 || open.PayAmount.ResourceOrEnergy() != PayResourceNone {
		t.Fatalf("opening prompt = %+v, want the opener's number from 0", open)
	}
	bid(t, g, a, 2)
	top := openPromptOfKind(g, PendingChoicePayAmount)
	if top == nil || top.PayAmount.ResourceOrEnergy() != PayResourceBid {
		t.Fatalf("second prompt = %+v, want a bid", top)
	}
	if err := g.ResolvePayAmount(top.ID, b.ID, 2); err == nil {
		t.Fatal("a bid equal to the high bid was taken; it must top it")
	}
	bid(t, g, b, 4)
	bid(t, g, a, 7)
	bid(t, g, b, 0)
	if auctionWinner != a.ID || auctionBid != 7 {
		t.Fatalf("winner %v at %d, want the opener at 7", auctionWinner, auctionBid)
	}
}

// A bidder who leaves the game passes (CR 800.4a).
func TestABidderWhoLeavesPasses(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	// The card asking is the opener's (a permanent standing in for the
	// resolving spell): a dropped prompt runs its default only while the
	// object it belongs to is still in the game (CR 800.4a).
	src := pushCombatant(t, g, a, "Auctioneer", 1, 1)
	startAuction(t, g, LifeAuction{Source: src, Bidders: []uuid.UUID{a.ID, b.ID, c.ID}, OpeningBid: 0, Label: "Illicit Auction"})
	bid(t, g, b, 2)
	if err := g.Concede(c.ID); err != nil {
		t.Fatalf("concede: %v", err)
	}
	if auctionBid < 0 {
		bid(t, g, a, 0)
	}
	if auctionWinner != b.ID || auctionBid != 2 {
		t.Fatalf("winner %v at %d, want %s at 2", auctionWinner, auctionBid, b.Name)
	}
}
