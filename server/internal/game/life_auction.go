package game

import (
	"strconv"

	"github.com/google/uuid"
)

// life_auction.go — life bidding in turn order (#6212): Pain's Reward,
// Mages' Contest, Illicit Auction.
//
//	"Each player may bid life. You start the bidding with a bid of any
//	number. In turn order, each player may top the high bid. The bidding
//	ends if the high bid stands. The high bidder loses life equal to the
//	high bid and …"
//
//	101.4. If multiple players would make choices and/or take actions at
//	       the same time, the active player … makes any choices required,
//	       then the next player in turn order …
//
// The CR has no bidding rule; the card's own text is the procedure. The
// bidders are listed in turn order starting with the player who opens
// (the card's "you"). The opening bid is either fixed by the card (Mages'
// Contest's 1, Illicit Auction's 0) or chosen by the opener (Pain's
// Reward's "any number": a number prompt with no ceiling). Then each
// other bidder in turn, skipping the high bidder, is asked to top the
// high bid or pass: a pay_amount prompt with Resource PayResourceBid,
// whose answers are 0 (pass) and anything above the high bid. A bid is
// not a payment (nothing is paid until the end), so there is no ceiling
// (NoMax). The bidding ends when every other bidder still in the game has
// passed in a row since the high bid was made. A bidder who has left the
// game (CR 800.4a) passes. The card's tail is then told the high bidder
// (IDs[0]) and the high bid (Amount).
//
// Every round is a registered number branch carrying the bidders and the
// state as data (Carry, Nums) and the card's tail (NumberBranch.Tail), so
// a table waiting on a bid is a restore point.

// PayResourceBid is a pay_amount prompt that asks a bidder to top the
// high bid: 0 is a pass, any number from Min up is a bid, and nothing is
// paid (#6212).
const PayResourceBid = "bid"

// PayAmountBid is the unit of a bid: one more life bid.
const PayAmountBid = "bid"

// LifeAuction is what StartLifeAuctionForEffect is handed.
type LifeAuction struct {
	// Source is the card holding the auction.
	Source uuid.UUID
	// Bidders are the players who may bid, in turn order, starting with
	// the opener (the card's "you"). At least two.
	Bidders []uuid.UUID
	// OpenerChooses makes the opening bid the opener's choice of any
	// number (Pain's Reward); otherwise it is OpeningBid.
	OpenerChooses bool
	OpeningBid    int
	// Label names the card in each prompt ("Illicit Auction").
	Label string
	// Then is the rest of the card: told the high bidder as IDs[0] and
	// the high bid as Amount.
	Then KeyedTail
}

var (
	lifeAuctionOpen NumberThen
	lifeAuctionBid  NumberThen
)

func init() {
	// On-disk keys: never renamed, never reused.
	lifeAuctionOpen = RegisterNumberThen("number/life-auction-open", lifeAuctionOpened)
	lifeAuctionBid = RegisterNumberThen("number/life-auction-bid", lifeAuctionBidAnswered)
}

// StartLifeAuctionForEffect runs the auction. Caller must hold g.mu
// (write).
func (g *Game) StartLifeAuctionForEffect(a LifeAuction) error {
	if len(a.Bidders) == 0 {
		return nil
	}
	label := a.Label
	if label == "" {
		label = "Bidding"
	}
	bidders := copyUUIDs(a.Bidders)
	if a.OpenerChooses && g.playerInGameLocked(bidders[0]) {
		return g.QueueChooseNumberForEffect(ChooseNumber{
			Chooser:  bidders[0],
			Source:   a.Source,
			Resource: PayResourceNone,
			Min:      0,
			NoMax:    true,
			Unit:     PayAmountBid,
			Question: label + ": start the bidding with a bid of any amount of life. The high bidder loses that much life.",
			Branch: NumberBranch{Then: lifeAuctionOpen, Carry: bidders, Words: []string{label},
				Tail: copyKeyedTail(a.Then)},
		})
	}
	opening := max(a.OpeningBid, 0)
	return g.askNextBidLocked(a.Source, label, bidders, opening, 0, 1, 0, a.Then)
}

func lifeAuctionOpened(g *Game, r NumberAnswer) error {
	return g.askNextBidLocked(r.Source, firstWord(r.Words), r.Carry, max(r.Number, 0), 0, 1, 0, r.Tail)
}

// askNextBidLocked asks the bidder at `cur` (or the next one still in the
// game who is not the high bidder) to top `high`, held by the bidder at
// `holder`; `passes` is how many have passed in a row since. With
// nobody left to ask, the auction ends.
//
// Caller must hold g.mu (write).
func (g *Game) askNextBidLocked(source uuid.UUID, label string, bidders []uuid.UUID, high, holder, cur, passes int, then KeyedTail) error {
	n := len(bidders)
	for {
		if passes >= g.otherBiddersLocked(bidders, holder) {
			winner := bidders[holder]
			if !g.playerInGameLocked(winner) {
				winner = uuid.Nil
			}
			return then.RunTold(g, []uuid.UUID{winner}, high)
		}
		cur %= n
		if cur == holder {
			cur++
			continue
		}
		bidder := bidders[cur]
		if !g.playerInGameLocked(bidder) {
			cur++
			continue
		}
		holderName := g.playerNameLocked(bidders[holder])
		return g.QueueChooseNumberForEffect(ChooseNumber{
			Chooser:  bidder,
			Source:   source,
			Resource: PayResourceBid,
			Min:      high + 1,
			NoMax:    true,
			Unit:     PayAmountBid,
			Question: label + ": the high bid is " + strconv.Itoa(high) + " life (" + holderName +
				"). Top it, or pass. The high bidder loses that much life.",
			Branch: NumberBranch{Then: lifeAuctionBid, Carry: copyUUIDs(bidders), Words: []string{label},
				Nums: []int{high, holder, cur, passes}, Tail: copyKeyedTail(then)},
		})
	}
}

// otherBiddersLocked is how many bidders other than the one at `holder`
// are still in the game: the passes that let the high bid stand.
//
// Caller must hold g.mu.
func (g *Game) otherBiddersLocked(bidders []uuid.UUID, holder int) int {
	n := 0
	for i, b := range bidders {
		if i != holder && g.playerInGameLocked(b) {
			n++
		}
	}
	return n
}

func lifeAuctionBidAnswered(g *Game, r NumberAnswer) error {
	if len(r.Nums) < 4 || len(r.Carry) == 0 {
		return nil
	}
	high, holder, cur, passes := r.Nums[0], r.Nums[1], r.Nums[2], r.Nums[3]
	label := firstWord(r.Words)
	// A pass, or a bidder who has left (CR 800.4a).
	if r.Number <= high || !g.playerInGameLocked(r.Chooser) {
		return g.askNextBidLocked(r.Source, label, r.Carry, high, holder, cur+1, passes+1, r.Tail)
	}
	return g.askNextBidLocked(r.Source, label, r.Carry, r.Number, cur, cur+1, 0, r.Tail)
}

// firstWord is words[0], or "".
func firstWord(words []string) string {
	if len(words) == 0 {
		return ""
	}
	return words[0]
}

// TurnOrderFromForEffect lists every player still in the game in turn
// order, starting with `first`: the bidders of "each player may bid".
// Caller must hold g.mu.
func (g *Game) TurnOrderFromForEffect(first uuid.UUID) []uuid.UUID {
	n := len(g.Seats)
	start := 0
	for i, s := range g.Seats {
		if s != nil && s.ID == first {
			start = i
		}
	}
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		s := g.Seats[(start+i)%n]
		if s != nil && !s.Eliminated {
			out = append(out, s.ID)
		}
	}
	return out
}
