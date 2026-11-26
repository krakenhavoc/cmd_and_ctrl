package testsl

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_bidding_test.go — #6212, LifeBidding: "You start the bidding with a
// bid of …. In turn order, each player may top the high bid. The bidding
// ends if the high bid stands." CR 101.4 ("in turn order"), CR 107.1b (a
// bid is a number with no printed ceiling), CR 119.3 (the high bidder
// loses that much life). The test branch does what Illicit Auction's and
// Pain's Reward's winners do with the result: lose the high bid in life.

var lifeBidTestWon = effects.NumberThen("number/test-life-bidding-winner-loses-the-bid", func(ctx *effects.Context, r game.NumberAnswer) error {
	if r.Chooser == uuid.Nil {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -r.Number)
})

func startLifeBidding(t *testing.T, g *game.Game, b effects.LifeBidding) {
	t.Helper()
	me := g.Seats[0]
	src := game.NewCard("Auction", me.ID)
	item := &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID, SourceCardID: src.InstanceID}
	var err error
	g.WithWriteLock(func() { err = b.Apply(effects.NewContext(g, item)) })
	if err != nil {
		t.Fatalf("LifeBidding: %v", err)
	}
}

func openBidChoice(t *testing.T, g *game.Game, kind game.PendingChoiceKind) *game.PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == kind {
			return c
		}
	}
	return nil
}

// stand lets the high bid stand; top bids n. Each checks who was asked.
func stand(t *testing.T, g *game.Game, who *game.Player) {
	t.Helper()
	c := openBidChoice(t, g, game.PendingChoiceOptionPick)
	if c == nil || c.Chooser != who.ID {
		t.Fatalf("want %s asked to top the bid, got %+v", who.Name, c)
	}
	if err := g.ResolveOptionPick(c.ID, who.ID, 0); err != nil {
		t.Fatalf("%s lets it stand: %v", who.Name, err)
	}
}

func top(t *testing.T, g *game.Game, who *game.Player, n int) {
	t.Helper()
	c := openBidChoice(t, g, game.PendingChoiceOptionPick)
	if c == nil || c.Chooser != who.ID {
		t.Fatalf("want %s asked to top the bid, got %+v", who.Name, c)
	}
	if err := g.ResolveOptionPick(c.ID, who.ID, 1); err != nil {
		t.Fatalf("%s tops it: %v", who.Name, err)
	}
	bid(t, g, who, n)
}

func bid(t *testing.T, g *game.Game, who *game.Player, n int) {
	t.Helper()
	c := openBidChoice(t, g, game.PendingChoicePayAmount)
	if c == nil || c.Chooser != who.ID {
		t.Fatalf("want %s asked for a bid, got %+v", who.Name, c)
	}
	if err := g.ResolvePayAmount(c.ID, who.ID, n); err != nil {
		t.Fatalf("%s bids %d: %v", who.Name, n, err)
	}
}

func lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

func noBidOpen(t *testing.T, g *game.Game) {
	t.Helper()
	for _, k := range []game.PendingChoiceKind{game.PendingChoiceOptionPick, game.PendingChoicePayAmount} {
		if c := openBidChoice(t, g, k); c != nil {
			t.Fatalf("the bidding is over but %s is still asked: %+v", k, c)
		}
	}
}

// The issue's acceptance: Illicit Auction's opening 0, bids 0→3→5, and the
// high bid stands once everyone else lets it: the 5-bidder loses 5 life.
func TestLifeBiddingTheHighBidStands(t *testing.T) {
	g := newBiddingGame(t)
	a, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	before := lives(g)
	startLifeBidding(t, g, effects.LifeBidding{Opening: 0, Then: lifeBidTestWon})
	top(t, g, b, 3)
	stand(t, g, c)
	top(t, g, d, 5)
	stand(t, g, a)
	stand(t, g, b)
	stand(t, g, c)
	noBidOpen(t, g)
	after := lives(g)
	for i := range g.Seats {
		want := before[i]
		if g.Seats[i] == d {
			want -= 5
		}
		if after[i] != want {
			t.Errorf("%s life = %d, want %d", g.Seats[i].Name, after[i], want)
		}
	}
}

// A player who let the bid stand may top it on a later pass round the
// table, and the round then starts again from them.
func TestLifeBiddingAPasserMayBidAgain(t *testing.T) {
	g := newBiddingGame(t)
	a, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	before := b.Life
	startLifeBidding(t, g, effects.LifeBidding{Opening: 0, Then: lifeBidTestWon})
	stand(t, g, b)
	stand(t, g, c)
	top(t, g, d, 1)
	stand(t, g, a)
	top(t, g, b, 2)
	stand(t, g, c)
	stand(t, g, d)
	stand(t, g, a)
	noBidOpen(t, g)
	if b.Life != before-2 {
		t.Errorf("%s life = %d, want %d", b.Name, b.Life, before-2)
	}
}

// "A bid of any number" (Pain's Reward): the controller is asked for the
// opening bid, and if nobody tops it they win it.
func TestLifeBiddingChosenOpening(t *testing.T) {
	g := newBiddingGame(t)
	a, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	before := a.Life
	startLifeBidding(t, g, effects.LifeBidding{OpeningChosen: true, Then: lifeBidTestWon})
	bid(t, g, a, 4)
	stand(t, g, b)
	stand(t, g, c)
	stand(t, g, d)
	noBidOpen(t, g)
	if a.Life != before-4 {
		t.Errorf("%s life = %d, want %d", a.Name, a.Life, before-4)
	}
}

// "You and target spell's controller bid life" (Mages' Contest): only the
// two bid, in turn order, and a top must be more than the high bid.
func TestLifeBiddingBetweenTwoPlayers(t *testing.T) {
	g := newBiddingGame(t)
	a, c := g.Seats[0], g.Seats[2]
	before := lives(g)
	startLifeBidding(t, g, effects.LifeBidding{
		Bidders: func(*effects.Context) []uuid.UUID { return []uuid.UUID{c.ID} },
		Opening: 1,
		Then:    lifeBidTestWon,
	})
	opt := openBidChoice(t, g, game.PendingChoiceOptionPick)
	if opt == nil || opt.Chooser != c.ID {
		t.Fatalf("want %s asked first, got %+v", c.Name, opt)
	}
	if err := g.ResolveOptionPick(opt.ID, c.ID, 1); err != nil {
		t.Fatal(err)
	}
	q := openBidChoice(t, g, game.PendingChoicePayAmount)
	if q == nil || q.PayAmount == nil || q.PayAmount.Min != 2 {
		t.Fatalf("the raise prompt = %+v, want a minimum of 2", q)
	}
	if err := g.ResolvePayAmount(q.ID, c.ID, 1); err == nil {
		t.Fatal("a bid that does not top the high bid was accepted")
	}
	bid(t, g, c, 2)
	top(t, g, a, 3)
	stand(t, g, c)
	noBidOpen(t, g)
	after := lives(g)
	for i := range g.Seats {
		want := before[i]
		if g.Seats[i] == a {
			want -= 3
		}
		if after[i] != want {
			t.Errorf("%s life = %d, want %d", g.Seats[i].Name, after[i], want)
		}
	}
}

// A table waiting on a bid is a restore point: the auction goes on in a
// game restored from a snapshot.
func TestLifeBiddingSurvivesASnapshot(t *testing.T) {
	g := newBiddingGame(t)
	startLifeBidding(t, g, effects.LifeBidding{Opening: 0, Then: lifeBidTestWon})
	top(t, g, g.Seats[1], 3)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap game.GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	b := r.Seats[1]
	before := b.Life
	stand(t, r, r.Seats[2])
	stand(t, r, r.Seats[3])
	stand(t, r, r.Seats[0])
	noBidOpen(t, r)
	if b.Life != before-3 {
		t.Errorf("restored: %s life = %d, want %d", b.Name, b.Life, before-3)
	}
}

func newBiddingGame(t *testing.T) *game.Game {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("active seat = %d, want 0", g.Turn.ActiveSeat)
	}
	return g
}
