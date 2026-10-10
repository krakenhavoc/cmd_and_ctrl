package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// choose_number_test.go — ADR 0129's amendment of 2026-10-09 (#1941):
// the pay_amount prompt asking for life, and for a number that is not
// paid.

func queueChooseNumber(t *testing.T, g *Game, p ChooseNumber) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueueChooseNumberForEffect(p); err != nil {
			t.Fatalf("QueueChooseNumberForEffect: %v", err)
		}
	})
}

// "Pay any amount of life": the ceiling is the payer's life total
// (CR 119.4), zero declines, an answer past the life total is refused,
// and a legal one is paid and handed on.
func TestChooseNumberPaysLifeUpToTheLifeTotal(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Life = 12
	got := -1
	queueChooseNumber(t, g, ChooseNumber{
		Chooser: me.ID, Source: uuid.New(), Resource: PayResourceLife, Goal: 2, Unit: PayAmountCards,
		Then: func(_ *Game, n int) error { got = n; return nil },
	})
	c := payAmountChoice(g)
	if c == nil || c.PayAmount == nil {
		t.Fatal("no pay_amount queued")
	}
	pa := c.PayAmount
	if pa.Max != 12 || pa.Min != 0 || pa.Goal != 2 || pa.ResourceOrEnergy() != PayResourceLife || pa.NoMax {
		t.Fatalf("prompt = %+v, want 0..12 life, goal 2", *pa)
	}
	for _, bad := range []int{-1, 13} {
		if err := g.ResolvePayAmount(c.ID, me.ID, bad); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("amount %d: err = %v, want ErrInvalidParam", bad, err)
		}
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 5); err != nil {
		t.Fatalf("resolve 5: %v", err)
	}
	if got != 5 || me.Life != 7 {
		t.Errorf("handed %d, life %d; want 5 and 7", got, me.Life)
	}
}

// A life payment may be declined, and a player with no life to pay, or
// whose life total can't change (CR 119.8), is not asked.
func TestChooseNumberLifeDeclinesAndSkipsAnUnpayablePlayer(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Life = 6
	got := -1
	queueChooseNumber(t, g, ChooseNumber{
		Chooser: me.ID, Source: uuid.New(), Resource: PayResourceLife,
		Then: func(_ *Game, n int) error { got = n; return nil },
	})
	if err := g.ResolvePayAmount(payAmountChoice(g).ID, me.ID, 0); err != nil {
		t.Fatalf("resolve 0: %v", err)
	}
	if got != 0 || me.Life != 6 {
		t.Errorf("declined: handed %d, life %d; want 0 and 6", got, me.Life)
	}

	me.Life = 0
	got = -1
	queueChooseNumber(t, g, ChooseNumber{
		Chooser: me.ID, Source: uuid.New(), Resource: PayResourceLife,
		Then: func(_ *Game, n int) error { got = n; return nil },
	})
	if got != 0 || payAmountChoice(g) != nil {
		t.Errorf("no life: handed %d, prompt %v; want 0 and none", got, payAmountChoice(g))
	}

	me.Life = 20
	lockUntilNextTurn(g, me)
	got = -1
	queueChooseNumber(t, g, ChooseNumber{
		Chooser: me.ID, Source: uuid.New(), Resource: PayResourceLife,
		Then: func(_ *Game, n int) error { got = n; return nil },
	})
	if got != 0 || payAmountChoice(g) != nil {
		t.Errorf("life total can't change: handed %d, prompt %v; want 0 and none", got, payAmountChoice(g))
	}
}

// A number that is not paid: no ceiling (CR 107.1b rules out negatives
// alone), so a very large answer is taken; the floor is the smallest
// answer and nothing is paid. Marks are kept in bounds and sorted.
func TestChooseNumberWithNoCeiling(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Life = 20
	got := -1
	queueChooseNumber(t, g, ChooseNumber{
		Chooser: me.ID, Source: uuid.New(), Resource: PayResourceNone, NoMax: true,
		Goal: 3, Marks: []int{20, 3, -2}, Unit: PayAmountDamage, SelfDamage: true,
		Then: func(_ *Game, n int) error { got = n; return nil },
	})
	c := payAmountChoice(g)
	if c == nil {
		t.Fatal("no prompt")
	}
	pa := c.PayAmount
	if !pa.NoMax || pa.Max != PayAmountNoMaxCeiling || pa.ResourceOrEnergy() != PayResourceNone || !pa.SelfDamage {
		t.Fatalf("prompt = %+v", *pa)
	}
	if len(pa.Marks) != 2 || pa.Marks[0] != 3 || pa.Marks[1] != 20 {
		t.Errorf("marks = %v, want [3 20]", pa.Marks)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, PayAmountNoMaxCeiling+1); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("past the guard: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 500); err != nil {
		t.Fatalf("resolve 500: %v", err)
	}
	if got != 500 || me.Life != 20 {
		t.Errorf("handed %d, life %d; want 500 and 20 (nothing paid)", got, me.Life)
	}
}

// A number that is not paid has no decline: with a floor of 1, zero is
// not an answer.
func TestChooseNumberFloorIsTheSmallestAnswer(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	queueChooseNumber(t, g, ChooseNumber{
		Chooser: me.ID, Source: uuid.New(), Resource: PayResourceNone, Min: 1, Max: 4,
	})
	c := payAmountChoice(g)
	if err := g.ResolvePayAmount(c.ID, me.ID, 0); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("zero under a floor of 1: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 4); err != nil {
		t.Errorf("resolve 4: %v", err)
	}
}

// The number stored on a permanent as it entered is cleared when it
// leaves the battlefield (CR 400.7), and remembered as its last-known
// information.
func TestChosenNumberIsClearedOnLeavingAndRemembered(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(Card{InstanceID: id, Name: "Processor", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { g.SetChosenNumberForEffect(id, 7) })
	if n := g.ChosenNumberOf(id); n != 7 {
		t.Fatalf("stored %d, want 7", n)
	}
	var info PermanentInfo
	g.WithWriteLock(func() { info = permanentInfoOf(findBattlefieldCard(g, id)) })
	if info.ChosenNumber != 7 {
		t.Errorf("PermanentInfo.ChosenNumber = %d, want 7", info.ChosenNumber)
	}
	if _, err := MoveCard(g.Battlefield, me.Graveyard, id); err != nil {
		t.Fatal(err)
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == id && c.ChosenNumber != 0 {
			t.Errorf("ChosenNumber %d survived leaving the battlefield", c.ChosenNumber)
		}
	}
}
