package game

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// energy_payment_test.go — ADR 0129 §3: energy paid while an ability
// resolves (CR 118.12, 118.12a, 107.14). A fixed amount is the pay-unless
// prompt with a PayActionEnergy payment; a chosen amount is pay_amount.

func queueEnergyPayment(t *testing.T, g *Game, p EnergyPayment) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueueEnergyPaymentForEffect(p); err != nil {
			t.Fatalf("QueueEnergyPaymentForEffect: %v", err)
		}
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePayUnless {
			return c.ID
		}
	}
	t.Fatal("no pay_unless queued")
	return uuid.Nil
}

// "You may pay {E}{E}. If you do": paid, the counters come off the payer
// with one negative player_counter_placed naming the source, and OnPay
// runs once.
func TestEnergyPaymentPaysAndRunsOnPay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	setEnergyForTest(t, g, me, 3)
	source := uuid.New()
	paid, declined := 0, 0
	id := queueEnergyPayment(t, g, EnergyPayment{
		Chooser: me.ID, Source: source, N: 2, Question: "pay {E}{E}?",
		OnPay:     func(*Game) error { paid++; return nil },
		OnDecline: func(*Game) error { declined++; return nil },
	})
	c := g.PendingChoices[0]
	if a := c.PayAction(); a == nil || a.Kind != PayActionEnergy || a.Count != 2 {
		t.Fatalf("PayAction = %+v, want energy ×2", a)
	}
	if c.PayCost != "{E}{E}" {
		t.Errorf("PayCost = %q, want {E}{E}", c.PayCost)
	}
	before := len(g.Events)
	if err := g.ResolvePayUnless(id, me.ID, true); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if paid != 1 || declined != 0 {
		t.Errorf("paid %d declined %d, want 1 and 0", paid, declined)
	}
	if PlayerEnergy(me) != 1 || me.Energy != 1 {
		t.Errorf("energy = %d (legacy %d), want 1", PlayerEnergy(me), me.Energy)
	}
	found := false
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventPlayerCounterPlaced && ev.Actor == me.ID && ev.Source == source {
			found = true
		}
	}
	if !found {
		t.Error("no player_counter_placed naming the payer and the source")
	}
}

// CR 118.3: a "yes" from a payer short of the energy pays nothing and is
// the decline, as an unfundable mana "yes" is.
func TestEnergyPaymentShortYesIsTheDecline(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	setEnergyForTest(t, g, me, 1)
	paid, declined := 0, 0
	id := queueEnergyPayment(t, g, EnergyPayment{
		Chooser: me.ID, Source: uuid.New(), N: 2,
		OnPay:     func(*Game) error { paid++; return nil },
		OnDecline: func(*Game) error { declined++; return nil },
	})
	if err := g.ResolvePayUnless(id, me.ID, true); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if paid != 0 || declined != 1 {
		t.Errorf("paid %d declined %d, want 0 and 1", paid, declined)
	}
	if PlayerEnergy(me) != 1 {
		t.Errorf("energy = %d, want 1 (nothing paid)", PlayerEnergy(me))
	}
}

// An energy payment names nothing: card IDs are refused and the prompt
// stays to be answered.
func TestEnergyPaymentRefusesCardIDs(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	setEnergyForTest(t, g, me, 2)
	id := queueEnergyPayment(t, g, EnergyPayment{Chooser: me.ID, Source: uuid.New(), N: 1})
	if err := g.ResolvePayUnlessWithCards(id, me.ID, true, []uuid.UUID{uuid.New()}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("err = %v, want ErrInvalidParam", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Error("a refused answer took the prompt away")
	}
}

// InThisStep anchors the prompt to the step, so it blocks the table
// there (#997's anchor).
func TestEnergyPaymentInThisStepIsOwed(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	setEnergyForTest(t, g, me, 2)
	queueEnergyPayment(t, g, EnergyPayment{Chooser: me.ID, Source: uuid.New(), N: 1, InThisStep: true})
	if !g.ChoicePromptBlocksTable(g.PendingChoices[0]) {
		t.Error("an InThisStep energy payment does not hold the step")
	}
}

// "Pay {E}{E}. If you can't": mandatory. A payer with the energy pays it;
// one without pays nothing and the "can't" branch runs.
func TestPayEnergyIfAble(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		have     int
		wantPaid bool
		wantLeft int
	}{{3, true, 1}, {1, false, 1}} {
		setEnergyForTest(t, g, me, tc.have)
		var got bool
		g.WithWriteLock(func() {
			if err := g.PayEnergyIfAbleForEffect(me.ID, uuid.New(), 2, func(_ *Game, paid bool) error {
				got = paid
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
		if got != tc.wantPaid || PlayerEnergy(me) != tc.wantLeft {
			t.Errorf("have %d: paid %v, left %d; want %v, %d", tc.have, got, PlayerEnergy(me), tc.wantPaid, tc.wantLeft)
		}
	}
}

func queuePayAmount(t *testing.T, g *Game, p PayEnergyAmount) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueuePayEnergyAmountForEffect(p); err != nil {
			t.Fatalf("QueuePayEnergyAmountForEffect: %v", err)
		}
	})
}

func payAmountChoice(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePayAmount {
			return c
		}
	}
	return nil
}

// "You may pay any amount of {E}": the prompt's ceiling is the payer's
// energy, it blocks the table, an answer outside 0 and Min..Max is
// refused, and a legal one pays and hands the amount on.
func TestPayAmountBoundsPaysAndHandsOn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	setEnergyForTest(t, g, me, 5)
	got := -1
	queuePayAmount(t, g, PayEnergyAmount{
		Chooser: me.ID, Source: uuid.New(), Min: 1, Goal: 3, Unit: PayAmountDamage,
		Then: func(_ *Game, n int) error { got = n; return nil },
	})
	c := payAmountChoice(g)
	if c == nil || c.PayAmount == nil {
		t.Fatal("no pay_amount queued")
	}
	if !reflect.DeepEqual(*c.PayAmount, PayAmountPrompt{Min: 1, Max: 5, Goal: 3, Unit: PayAmountDamage, Resource: PayResourceEnergy}) {
		t.Errorf("PayAmount = %+v", *c.PayAmount)
	}
	if !ChoiceBlocksTable(PendingChoicePayAmount) {
		t.Error("pay_amount does not block the table")
	}
	for _, bad := range []int{-1, 6} {
		if err := g.ResolvePayAmount(c.ID, me.ID, bad); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("amount %d: err = %v, want ErrInvalidParam", bad, err)
		}
	}
	if err := g.ResolvePayAmount(c.ID, g.Seats[1].ID, 2); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("wrong chooser: err = %v", err)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 4); err != nil {
		t.Fatalf("resolve 4: %v", err)
	}
	if got != 4 || PlayerEnergy(me) != 1 {
		t.Errorf("handed %d, energy %d; want 4 and 1", got, PlayerEnergy(me))
	}
}

// Zero is always an answer: "one or more" may be declined.
func TestPayAmountZeroDeclines(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	setEnergyForTest(t, g, me, 2)
	got := -1
	queuePayAmount(t, g, PayEnergyAmount{
		Chooser: me.ID, Source: uuid.New(), Min: 1,
		Then: func(_ *Game, n int) error { got = n; return nil },
	})
	c := payAmountChoice(g)
	if err := g.ResolvePayAmount(c.ID, me.ID, 0); err != nil {
		t.Fatalf("resolve 0: %v", err)
	}
	if got != 0 || PlayerEnergy(me) != 2 {
		t.Errorf("handed %d, energy %d; want 0 and 2", got, PlayerEnergy(me))
	}
}

// Nothing to choose (no energy, or less than "one or more" needs): no
// prompt, and Then runs at once with 0. A negative Goal is "as much as
// you can".
func TestPayAmountNothingToChoose(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	setEnergyForTest(t, g, me, 0)
	got := -1
	queuePayAmount(t, g, PayEnergyAmount{
		Chooser: me.ID, Source: uuid.New(),
		Then: func(_ *Game, n int) error { got = n; return nil },
	})
	if got != 0 || payAmountChoice(g) != nil {
		t.Errorf("handed %d, prompt %v; want 0 and none", got, payAmountChoice(g))
	}
	setEnergyForTest(t, g, me, 4)
	queuePayAmount(t, g, PayEnergyAmount{Chooser: me.ID, Source: uuid.New(), Goal: -1})
	if c := payAmountChoice(g); c == nil || c.PayAmount.Goal != 4 {
		t.Errorf("a negative goal is not the ceiling: %+v", c)
	}
}

// CR 800.4f: a departed payer's energy is not paid. The rest of the
// card runs with nothing paid, and nobody else is asked.
func TestPayAmountDroppedRunsWithNothingPaid(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner, leaver := g.Seats[0], g.Seats[1]
	sourceID := departureTestSource(g, owner.ID, "Harnessed Lightning")
	setEnergyForTest(t, g, leaver, 3)
	got, runs := -1, 0
	queuePayAmount(t, g, PayEnergyAmount{
		Chooser: leaver.ID, Source: sourceID,
		Then: func(_ *Game, n int) error { got = n; runs++; return nil },
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if runs != 1 || got != 0 {
		t.Errorf("Then ran %d times with %d, want once with 0", runs, got)
	}
	if payAmountChoice(g) != nil {
		t.Error("the prompt was reassigned")
	}
}
