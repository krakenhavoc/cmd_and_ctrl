package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// any_player_pays_test.go — ADR 0150 decisions 1 and 2 (#1903): one
// payment offered to the whole table, one pay-unless at a time, in
// APNAP order from the active player, stopping at the first payer.

type offerOutcome struct {
	paidBy []uuid.UUID
	nobody int
}

// queueOffer makes `controller`'s offer of `cost` straight at the
// engine, so the rule is measured without a card in the way.
func queueOffer(t *testing.T, g *Game, controller uuid.UUID, cost string, out *offerOutcome, mutate ...func(*AnyPlayerPays)) {
	t.Helper()
	source := departureTestSource(g, controller, "Rhystic Thing")
	p := AnyPlayerPays{
		Source:     source,
		Controller: controller,
		Cost:       cost,
		Question:   "Rhystic Thing — pay " + cost + "?",
		OnPaid: func(_ *Game, payer uuid.UUID) error {
			out.paidBy = append(out.paidBy, payer)
			return nil
		},
		OnNobody: func(*Game) error {
			out.nobody++
			return nil
		},
	}
	for _, m := range mutate {
		m(&p)
	}
	g.WithWriteLock(func() {
		if err := g.QueueAnyPlayerPaysForEffect(p); err != nil {
			t.Fatalf("QueueAnyPlayerPaysForEffect: %v", err)
		}
	})
}

// answerOffer answers the one open link as its chooser.
func answerOffer(t *testing.T, g *Game, pay bool) uuid.UUID {
	t.Helper()
	c := onlyPendingChoice(t, g)
	if err := g.ResolvePayUnless(c.ID, c.Chooser, pay); err != nil {
		t.Fatalf("ResolvePayUnless(%v): %v", pay, err)
	}
	return c.Chooser
}

func giveColorless(p *Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(ManaToken{Color: "C"})
	}
}

func seatIDs(g *Game, idx ...int) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(idx))
	for _, i := range idx {
		out = append(out, g.Seats[i].ID)
	}
	return out
}

// The order is APNAP from the ACTIVE player whoever controls the
// source, the controller included (owner decision 1), and everyone
// declining runs OnNobody exactly once.
func TestAnyPlayerOfferAsksEveryPlayerInAPNAPOrderFromTheActivePlayer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	// Seat 0 is active; the effect is seat 2's.
	controller := g.Seats[2].ID
	var out offerOutcome
	queueOffer(t, g, controller, "{2}", &out)

	var asked []uuid.UUID
	for i := 0; i < 4; i++ {
		c := onlyPendingChoice(t, g)
		if c.Kind != PendingChoicePayUnless {
			t.Fatalf("link %d is a %q, want pay_unless", i, c.Kind)
		}
		o := c.AnyPlayerOffer
		if o == nil || o.Controller != controller || !sameIDs(o.Order, seatIDs(g, 0, 1, 2, 3)) {
			t.Fatalf("link %d carries offer %+v", i, o)
		}
		if !sameIDs(o.Declined, asked) {
			t.Fatalf("link %d: declined = %v, want %v", i, o.Declined, asked)
		}
		asked = append(asked, answerOffer(t, g, false))
	}
	if !sameIDs(asked, seatIDs(g, 0, 1, 2, 3)) {
		t.Errorf("asked %v, want seats 0..3 in order", asked)
	}
	if out.nobody != 1 || len(out.paidBy) != 0 {
		t.Errorf("nobody ran %d times, paid by %v; want once and nobody", out.nobody, out.paidBy)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left after the last decline", len(g.PendingChoices))
	}
	declines := 0
	for _, ev := range g.Events {
		if ev.Kind == EventAnyPlayerDeclined {
			declines++
		}
	}
	if declines != 4 {
		t.Errorf("%d decline events, want 4", declines)
	}
}

// The first payer ends the offer: OnPaid gets them, OnNobody never
// runs and nobody after them is asked.
func TestTheFirstPayerEndsTheAnyPlayerOffer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	controller := g.Seats[0].ID
	giveColorless(g.Seats[1], 2)
	giveColorless(g.Seats[2], 2)
	var out offerOutcome
	queueOffer(t, g, controller, "{2}", &out)

	answerOffer(t, g, false) // seat 0, the controller
	payer := answerOffer(t, g, true)
	if payer != g.Seats[1].ID {
		t.Fatalf("the second link went to %v, want seat 1", payer)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("the offer went on after seat 1 paid: %d prompts", len(g.PendingChoices))
	}
	if !sameIDs(out.paidBy, seatIDs(g, 1)) || out.nobody != 0 {
		t.Errorf("paid by %v, nobody %d; want seat 1 and no OnNobody", out.paidBy, out.nobody)
	}
	if n := len(g.Seats[1].ManaPool); n != 0 {
		t.Errorf("seat 1 has %d mana left, want the {2} spent", n)
	}
	if n := len(g.Seats[2].ManaPool); n != 2 {
		t.Errorf("seat 2 paid too: %d mana left", n)
	}
	paid := false
	for _, ev := range g.Events {
		if ev.Kind == EventAnyPlayerPaid && ev.Actor == payer && ev.Label == "{2}" {
			paid = true
		}
	}
	if !paid {
		t.Error("no EventAnyPlayerPaid for seat 1")
	}
}

// A "Pay" the offeree cannot fund is a decline (CR 118.3) and the
// offer moves on.
func TestAnUnfundablePayMovesTheAnyPlayerOfferOn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	for _, s := range g.Seats {
		s.ManaPool = nil
	}
	// Nobody has lands in this fixture's battlefield, so the auto-tapper
	// finds nothing either.
	var out offerOutcome
	queueOffer(t, g, g.Seats[0].ID, "{2}", &out)
	answerOffer(t, g, true)
	c := onlyPendingChoice(t, g)
	if c.Chooser != g.Seats[1].ID || !sameIDs(c.AnyPlayerOffer.Declined, seatIDs(g, 0)) {
		t.Fatalf("after an unfundable pay the offer is at %v with declined %v", c.Chooser, c.AnyPlayerOffer.Declined)
	}
	if len(out.paidBy) != 0 {
		t.Errorf("an unfundable pay counted as paid: %v", out.paidBy)
	}
}

// Every link blocks the step it was asked in and takes no standing
// answer (owner decision 2).
func TestAnAnyPlayerLinkBlocksItsStepAndTakesNoStandingAnswer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	var out offerOutcome
	queueOffer(t, g, g.Seats[0].ID, "{1}", &out)
	for i := 0; i < 2; i++ {
		c := onlyPendingChoice(t, g)
		if want := (TurnStep{Turn: g.Turn.Seq, Step: g.Turn.Step}); c.OwedInStep != want {
			t.Fatalf("link %d owed in %v, want %v", i, c.OwedInStep, want)
		}
		if !g.ChoicePromptBlocksTable(c) {
			t.Fatalf("link %d does not block the table", i)
		}
		if c.AutoAnswerKey != "" || payUnlessCovered(c) {
			t.Fatalf("link %d takes a standing answer (key %q)", i, c.AutoAnswerKey)
		}
		if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
			t.Fatalf("pass_priority with link %d open = %v, want ErrChoicePending", i, err)
		}
		answerOffer(t, g, false)
	}
}

// Guards holds the spell the outcome decides on the stack.
func TestAnAnyPlayerOfferGuardsTheSpellItDecides(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	spell := uuid.New()
	var out offerOutcome
	queueOffer(t, g, g.Seats[0].ID, "{1}", &out, func(p *AnyPlayerPays) { p.Guards = spell })
	c := onlyPendingChoice(t, g)
	if c.GuardsStackItem != spell {
		t.Errorf("the link guards %v, want %v", c.GuardsStackItem, spell)
	}
}

// Opponents skips the controller.
func TestAnAnyPlayerOfferToOpponentsSkipsTheController(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	var out offerOutcome
	queueOffer(t, g, g.Seats[0].ID, "{1}", &out, func(p *AnyPlayerPays) { p.Opponents = true })
	c := onlyPendingChoice(t, g)
	if c.Chooser != g.Seats[1].ID || !sameIDs(c.AnyPlayerOffer.Order, seatIDs(g, 1, 2, 3)) || !c.AnyPlayerOffer.Opponents {
		t.Fatalf("an any-opponent offer asks %v with order %v", c.Chooser, c.AnyPlayerOffer.Order)
	}
}

// A player who left before their turn is skipped, and one who leaves
// while asked declines and the offer moves on (CR 800.4f).
func TestDepartedPlayersAreSkippedOrDeclineTheAnyPlayerOffer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if err := g.Concede(g.Seats[2].ID); err != nil {
		t.Fatalf("Concede seat 2: %v", err)
	}
	var out offerOutcome
	queueOffer(t, g, g.Seats[0].ID, "{1}", &out)
	if c := onlyPendingChoice(t, g); !sameIDs(c.AnyPlayerOffer.Order, seatIDs(g, 0, 1, 3)) {
		t.Fatalf("order %v includes a player who left", c.AnyPlayerOffer.Order)
	}
	answerOffer(t, g, false) // seat 0
	if c := onlyPendingChoice(t, g); c.Chooser != g.Seats[1].ID {
		t.Fatalf("second link to %v, want seat 1", c.Chooser)
	}
	// Seat 1 leaves with the link open: a decline, so seat 3 is asked.
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatalf("Concede seat 1: %v", err)
	}
	c := onlyPendingChoice(t, g)
	if c.Chooser != g.Seats[3].ID || !sameIDs(c.AnyPlayerOffer.Declined, seatIDs(g, 0, 1)) {
		t.Fatalf("after seat 1 left the offer is at %v, declined %v", c.Chooser, c.AnyPlayerOffer.Declined)
	}
	answerOffer(t, g, false)
	if out.nobody != 1 {
		t.Errorf("OnNobody ran %d times, want 1", out.nobody)
	}
}

// The offer is plain data on the prompt and survives the undo clone
// deep-copied.
func TestTheAnyPlayerOfferIsClonedNotShared(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	var out offerOutcome
	queueOffer(t, g, g.Seats[0].ID, "{1}", &out)
	clone := g.Clone()
	orig := onlyPendingChoice(t, g).AnyPlayerOffer
	cp := clone.PendingChoices[0].AnyPlayerOffer
	if cp == nil || cp == orig || !sameIDs(cp.Order, orig.Order) {
		t.Fatalf("clone offer %+v vs %+v", cp, orig)
	}
	cp.Order[0] = uuid.Nil
	if orig.Order[0] == uuid.Nil {
		t.Error("the clone shares the offer's order with the original")
	}
}

// An unparseable cost is one error and the effect, not one per player.
func TestAnUnparseableAnyPlayerCostRunsTheEffectOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	var out offerOutcome
	queueOffer(t, g, g.Seats[0].ID, "{nope", &out)
	if len(g.PendingChoices) != 0 || out.nobody != 1 {
		t.Errorf("prompts %d, nobody %d; want none and once", len(g.PendingChoices), out.nobody)
	}
}
