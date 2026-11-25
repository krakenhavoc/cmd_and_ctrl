package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_bidding.go — "Each player may bid life. You start the bidding with
// a bid of …. In turn order, each player may top the high bid. The bidding
// ends if the high bid stands." (Pain's Reward, Mages' Contest, Illicit
// Auction; #6212, ADR 0129 amendment 2026-11-25).
//
// The Comprehensive Rules have no bidding rule; the card text is the
// procedure. "In turn order" is CR 101.4's order, starting from the
// player who started the bidding. Each bidder in turn, beginning with the
// one after the high bidder, either lets the high bid stand or tops it
// with a higher number; a player who let it stand may top it on a later
// pass round the table. The bidding ends when the turn comes back round to
// the high bidder with nobody having topped them: the high bid stands. A
// bid is a number, not a payment (CR 107.1b: no printed ceiling), so a
// player may bid more life than they have; the winner then loses that
// much life (CR 119.3) and the state-based actions do the rest.
//
// No new prompt: each step is an existing keyed prompt the client and the
// bot already answer, an option_pick ("Let the high bid stand" / "Top the
// high bid") and then a pay_amount number of at least one more than the
// high bid. The whole state of the auction rides each prompt as data
// (Carry: the bidders in turn order; Nums: the high bid, the high
// bidder's place and the place being asked; Words: the card's branch), so
// a table waiting on a bid is a restore point.

// LifeBidding runs the auction, then the card's registered number branch
// (Then, made with NumberThen) with the high bidder as r.Chooser and the
// high bid as r.Number, against the resolving item (its targets are read
// there, as ctx.Item). r.Chooser is uuid.Nil when the high bidder has left
// the game (CR 800.4a).
type LifeBidding struct {
	// Bidders are the players who bid ("you and target spell's
	// controller"); nil is every player still in the game. The resolving
	// effect's controller always bids and starts.
	Bidders func(ctx *Context) []uuid.UUID
	// Opening is the controller's opening bid ("a bid of 0", "of 1").
	// With OpeningChosen the controller is asked for it instead ("a bid
	// of any number"), from 0 up.
	Opening       int
	OpeningChosen bool
	// Then is what the high bidder's win does.
	Then game.NumberThen
}

// Nums layout of every bidding prompt.
const (
	bidHigh = iota
	bidHighAt
	bidAsking
	bidNums
)

// Apply asks the opening bid, or records the printed one, and starts the
// pass round the table.
func (b LifeBidding) Apply(ctx *Context) error {
	if b.Then.Key() == "" {
		return fmt.Errorf("%w: LifeBidding with no Then", game.ErrInvalidParam)
	}
	bidders := b.biddersInTurnOrder(ctx)
	words := []string{b.Then.Key()}
	if b.OpeningChosen {
		return ChooseNumber{
			Chooser:    ctx.Controller(),
			Min:        0,
			NoMax:      true,
			Question:   lifeBidCardName(ctx) + " — you start the bidding: bid any amount of life",
			Unit:       game.PayAmountOther,
			SelfDamage: true,
			Then:       lifeBidOpened,
			Carry:      bidders,
			Words:      words,
			Nums:       []int{0, 0, 1},
		}.Apply(ctx)
	}
	return lifeBidStep(ctx, bidders, words, []int{b.Opening, 0, 1})
}

// biddersInTurnOrder is the bidders, the controller first and then each
// other one in turn order (CR 101.4), leaving out anyone who has left the
// game.
func (b LifeBidding) biddersInTurnOrder(ctx *Context) []uuid.UUID {
	g := ctx.Game
	me := ctx.Controller()
	allowed := map[uuid.UUID]bool{me: true}
	if b.Bidders != nil {
		for _, id := range b.Bidders(ctx) {
			allowed[id] = true
		}
	}
	from := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == me {
			from = i
		}
	}
	out := []uuid.UUID{me}
	if from < 0 {
		return out
	}
	for k := 1; k < len(g.Seats); k++ {
		p := g.Seats[g.SeatInTurnOrderForEffect(from, k)]
		if p == nil || p.Eliminated || (b.Bidders != nil && !allowed[p.ID]) {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

func lifeBidCardName(ctx *Context) string {
	if c, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok && c.Name != "" {
		return c.Name
	}
	return "Bidding"
}

func bidderStillIn(g *game.Game, id uuid.UUID) bool {
	p := g.PlayerByIDForEffect(id)
	return p != nil && !p.Eliminated
}

// lifeBidStep asks the next bidder after the high bidder who is still in
// the game, or, once the turn has come back round to the high bidder, ends
// the bidding.
func lifeBidStep(ctx *Context, bidders []uuid.UUID, words []string, nums []int) error {
	n := len(bidders)
	if n == 0 || len(nums) < bidNums || len(words) == 0 {
		return fmt.Errorf("%w: a life bid with no state", game.ErrInvalidParam)
	}
	high, highAt, asking := nums[bidHigh], nums[bidHighAt], nums[bidAsking]%n
	for asking != highAt {
		who := bidders[asking]
		if bidderStillIn(ctx.Game, who) {
			holder := "you"
			if bidders[highAt] != who {
				holder = lifeBidPlayerName(ctx.Game, bidders[highAt])
			}
			return PickOption{
				Player:   who,
				Question: fmt.Sprintf("%s — the high bid is %d life (%s). Top it?", lifeBidCardName(ctx), high, holder),
				Options: []game.ChoiceOption{
					{Label: fmt.Sprintf("Let the high bid of %d life stand", high)},
					{Label: fmt.Sprintf("Top the high bid (bid %d or more)", high+1)},
				},
				ThenKey:  lifeBidPicked,
				Carry:    bidders,
				WithItem: true,
				Words:    words,
				Nums:     []int{high, highAt, asking},
			}.Apply(ctx)
		}
		asking = (asking + 1) % n
	}
	then, ok := game.NumberThenOf(words[0])
	if !ok {
		return fmt.Errorf("%w: life bidding branch %q", game.ErrUnknownEffectKey, words[0])
	}
	winner := bidders[highAt]
	if !bidderStillIn(ctx.Game, winner) {
		winner = uuid.Nil
	}
	return ctx.Game.RunNumberThenForEffect(game.NumberBranch{
		Then:  then,
		Carry: []uuid.UUID{ctx.Controller()},
		Asked: game.NewAskedItem(ctx.Item),
	}, winner, ctx.Source(), high)
}

func lifeBidPlayerName(g *game.Game, id uuid.UUID) string {
	if p := g.PlayerByIDForEffect(id); p != nil && p.Name != "" {
		return p.Name
	}
	return "another player"
}

// lifeBidOpened records the controller's chosen opening bid.
func lifeBidOpen(ctx *Context, r game.NumberAnswer) error {
	nums := append([]int(nil), r.Nums...)
	if len(nums) < bidNums {
		return fmt.Errorf("%w: a life bid with no state", game.ErrInvalidParam)
	}
	nums[bidHigh] = r.Number
	return lifeBids.step(ctx, r.Carry, r.Words, nums)
}

// lifeBidPicked is a bidder's answer: let it stand (or no answer: the
// chooser left), and the next bidder is asked; top it, and they name
// their bid.
func lifeBidAnswer(ctx *Context, r game.OptionPicked) error {
	nums := append([]int(nil), r.Nums...)
	if len(nums) < bidNums || len(r.Carry) == 0 {
		return fmt.Errorf("%w: a life bid with no state", game.ErrInvalidParam)
	}
	if r.Index != 1 {
		nums[bidAsking] = (nums[bidAsking] + 1) % len(r.Carry)
		return lifeBids.step(ctx, r.Carry, r.Words, nums)
	}
	return ChooseNumber{
		Chooser:    r.Chooser,
		Min:        nums[bidHigh] + 1,
		NoMax:      true,
		Question:   fmt.Sprintf("%s — your bid (more than %d life)", lifeBidCardName(ctx), nums[bidHigh]),
		Unit:       game.PayAmountOther,
		SelfDamage: true,
		Goal:       func(*Context) int { return nums[bidHigh] + 1 },
		Then:       lifeBidRaised,
		Carry:      r.Carry,
		Words:      r.Words,
		Nums:       nums,
	}.Apply(ctx)
}

// lifeBidRaised makes the bidder who topped the bid the high bidder, and
// the bidding goes on round the table from the next player.
func lifeBidRaise(ctx *Context, r game.NumberAnswer) error {
	nums := append([]int(nil), r.Nums...)
	if len(nums) < bidNums || len(r.Carry) == 0 {
		return fmt.Errorf("%w: a life bid with no state", game.ErrInvalidParam)
	}
	nums[bidHigh] = r.Number
	nums[bidHighAt] = nums[bidAsking] % len(r.Carry)
	nums[bidAsking] = (nums[bidHighAt] + 1) % len(r.Carry)
	return lifeBids.step(ctx, r.Carry, r.Words, nums)
}

// The three hops' on-disk keys. Each hop goes on through lifeBids, an
// interface, rather than calling lifeBidStep: a call through an interface
// is not an initialisation reference, so the keys can be var initialisers
// (a core file holds no init()) though the step refers to them.
var (
	lifeBidOpened = NumberThen("number/life-bidding-opening-bid", lifeBidOpen)
	lifeBidPicked = OptionPickThen("option-pick/life-bidding-top-or-stand", lifeBidAnswer)
	lifeBidRaised = NumberThen("number/life-bidding-raise", lifeBidRaise)
)

type lifeBidStepper interface {
	step(ctx *Context, bidders []uuid.UUID, words []string, nums []int) error
}

type lifeBidSteps struct{}

func (lifeBidSteps) step(ctx *Context, bidders []uuid.UUID, words []string, nums []int) error {
	return lifeBidStep(ctx, bidders, words, nums)
}

var lifeBids lifeBidStepper = lifeBidSteps{}
