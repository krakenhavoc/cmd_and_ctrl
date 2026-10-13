package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// any_player_payments_test.go — ADR 0150 §3 (#1903, PR 2): an
// any-player offer paid with life, with damage, or with nothing, and a
// price worked out for each player as they are asked.

func withPayment(a PayAction, words string) func(*AnyPlayerPays) {
	return func(p *AnyPlayerPays) {
		p.Action = &a
		p.Cost = words
	}
}

// A life payment is a payment only from a life total of at least the
// amount (CR 119.4): a short player's "Pay" is a decline and moves the
// offer on with their life untouched, and the next player pays.
func TestALifePaymentNeedsTheLifeAndMovesOnWithout(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	short, payer := g.Seats[0], g.Seats[1]
	short.Life, payer.Life = 4, 20
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "5 life", &out, withPayment(PayAction{Kind: PayActionLife, Count: 5}, "5 life"))

	if c := onlyPendingChoice(t, g); c.PayAction() == nil || c.PayAction().Kind != PayActionLife {
		t.Fatalf("the link's payment is %+v, want 5 life", c.PayAction())
	}
	if got := answerOffer(t, g, true); got != short.ID {
		t.Fatalf("asked %v first, want the active player", got)
	}
	if short.Life != 4 {
		t.Errorf("the short player's life went %d → %d", 4, short.Life)
	}
	if got := answerOffer(t, g, true); got != payer.ID {
		t.Fatalf("asked %v second, want seat 1", got)
	}
	if payer.Life != 15 {
		t.Errorf("payer's life is %d, want 15", payer.Life)
	}
	if !sameIDs(out.paidBy, []uuid.UUID{payer.ID}) || out.nobody != 0 {
		t.Errorf("paid by %v, nobody %d; want seat 1 and never", out.paidBy, out.nobody)
	}
}

// CR 119.8: a player whose life total can't change can't pay life at
// all, whatever their life total.
func TestALifePaymentIsRefusedWhileTheLifeTotalCantChange(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	lockUntilNextTurn(g, me)
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "5 life", &out, withPayment(PayAction{Kind: PayActionLife, Count: 5}, "5 life"))
	before := me.Life
	answerOffer(t, g, true)
	if me.Life != before || len(out.paidBy) != 0 {
		t.Errorf("a locked life total paid: life %d → %d, paid by %v", before, me.Life, out.paidBy)
	}
	if c := onlyPendingChoice(t, g); c.Chooser != g.Seats[1].ID {
		t.Errorf("the offer did not move on: asking %v", c.Chooser)
	}
}

// Temporal Extortion's price is each player's own half, rounded up,
// worked out as their link is queued (ADR 0150 §3).
func TestAPerPayerPriceIsEachPlayersOwnHalfRoundedUp(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.Seats[0].Life, g.Seats[1].Life = 40, 7
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "", &out, func(p *AnyPlayerPays) {
		p.PerPayer = func(g *Game, payer uuid.UUID) (PayAction, string) {
			a := PayAction{Kind: PayActionLife, Count: HalfLifeRoundedUp(g.playerByIDLocked(payer).Life)}
			return a, PayActionWords(a)
		}
	})
	first := onlyPendingChoice(t, g)
	if first.PayCost != "20 life" || first.PayAction().Count != 20 {
		t.Fatalf("seat 0's price is %q (%d), want 20 life", first.PayCost, first.PayAction().Count)
	}
	answerOffer(t, g, false)
	second := onlyPendingChoice(t, g)
	if second.PayCost != "4 life" || second.PayAction().Count != 4 {
		t.Fatalf("seat 1's price is %q (%d), want 4 life (7 halved, rounded up)", second.PayCost, second.PayAction().Count)
	}
	answerOffer(t, g, true)
	if g.Seats[1].Life != 3 || !sameIDs(out.paidBy, seatIDs(g, 1)) {
		t.Errorf("seat 1 at %d life, paid by %v; want 3 and seat 1", g.Seats[1].Life, out.paidBy)
	}
	var labels []string
	for _, ev := range g.Events {
		if ev.Kind == EventAnyPlayerDeclined || ev.Kind == EventAnyPlayerPaid {
			labels = append(labels, ev.Label)
		}
	}
	if len(labels) != 2 || labels[0] != "20 life" || labels[1] != "4 life" {
		t.Errorf("logged %v, want each player's own price", labels)
	}
}

// Taking damage is always a payment: the source deals it to the payer
// through the ordinary path, then the "yes" branch runs.
func TestADamagePaymentIsDealtByTheSourceAndAlwaysPays(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	me.Life = 3
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "5 damage", &out, withPayment(PayAction{Kind: PayActionDamage, Count: 5}, "5 damage"),
		func(p *AnyPlayerPays) { p.AcceptLabel, p.DeclineLabel = "Take 5 damage", "Don't" })
	c := onlyPendingChoice(t, g)
	if c.AcceptLabel != "Take 5 damage" || c.DeclineLabel != "Don't" {
		t.Errorf("labels %q / %q, want the card's words", c.AcceptLabel, c.DeclineLabel)
	}
	source := c.Source
	answerOffer(t, g, true)
	if me.Life != -2 {
		t.Errorf("life %d, want 3 - 5 = -2 (damage is not limited by the life total)", me.Life)
	}
	if !sameIDs(out.paidBy, []uuid.UUID{me.ID}) {
		t.Errorf("paid by %v, want the active player", out.paidBy)
	}
	dealt := false
	for _, ev := range g.Events {
		if ev.Kind == EventDealDamage && ev.Source == source && ev.Amount == 5 {
			dealt = true
		}
	}
	if !dealt {
		t.Error("no 5-damage event from the offer's source")
	}
}

// An offer that costs nothing (Distant Memories) is always payable, and
// a payment that names nothing refuses card IDs with the prompt kept.
func TestANoPaymentOfferAlwaysPaysAndRefusesCardIDs(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	var out offerOutcome
	queueOffer(t, g, g.Seats[0].ID, "", &out, withPayment(PayAction{Kind: PayActionNothing}, ""),
		func(p *AnyPlayerPays) { p.Opponents = true })
	c := onlyPendingChoice(t, g)
	if c.Chooser != g.Seats[1].ID {
		t.Fatalf("asked %v first, want seat 1 (the controller is skipped)", c.Chooser)
	}
	err := g.ResolvePayUnlessWithCards(c.ID, c.Chooser, true, []uuid.UUID{uuid.New()})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("card IDs on a no-payment answer: %v, want ErrInvalidParam", err)
	}
	if onlyPendingChoice(t, g).ID != c.ID {
		t.Fatal("the refused answer took the prompt")
	}
	answerOffer(t, g, true)
	if !sameIDs(out.paidBy, seatIDs(g, 1)) || out.nobody != 0 {
		t.Errorf("paid by %v, nobody %d; want seat 1", out.paidBy, out.nobody)
	}
}
