package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// either_payment_test.go — ADR 0187 §5 (#3575): an any-player offer with
// two ways to pay (Worms of the Earth: sacrifice two lands, or have it
// deal 5 damage). One question, and the answer names the branch.
//
// No catalog card uses it until Worms (slice 8), so the offer is queued
// straight at the engine, as any_player_payments_test.go does.

// wormsPayment is Worms of the Earth's payment: branch 0 sacrifices two
// lands, branch 1 takes 5 damage.
func wormsPayment() func(*anyPlayerTestOffer) {
	return withPayment(PayAction{
		Kind: PayActionEither,
		Either: []PayAction{
			{Kind: PayActionSacrifice, Count: 2, Of: PermanentQuery{Types: []string{"land"}}, Words: "sacrifice two lands"},
			{Kind: PayActionDamage, Count: 5, Words: "take 5 damage"},
		},
	}, "sacrifice two lands or take 5 damage")
}

// landsFor puts n basic lands on p's battlefield and returns their IDs.
func landsFor(g *Game, p *Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		id := uuid.New()
		g.Battlefield.PushTop(Card{
			InstanceID: id,
			Name:       "Forest",
			TypeLine:   "Basic Land — Forest",
			Owner:      p.ID,
			Controller: p.ID,
		})
		out = append(out, id)
	}
	return out
}

func intp(n int) *int { return &n }

// A pay_option out of range, a pay_option on a prompt that is not an
// either payment, and apply:true without a pay_option are each refused
// with the prompt kept, and nothing is paid.
func TestEitherPaymentRefusesAMalformedAnswerAndKeepsThePrompt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	lands := landsFor(g, me, 3)
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "either", &out, wormsPayment())
	c := onlyPendingChoice(t, g)

	cases := map[string]func() error{
		"apply without a pay_option": func() error {
			return g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, nil, nil)
		},
		"apply without a pay_option, with cards": func() error {
			return g.ResolvePayUnlessWithCards(c.ID, c.Chooser, true, lands[:2])
		},
		"apply through the plain resolver": func() error {
			return g.ResolvePayUnless(c.ID, c.Chooser, true)
		},
		"a pay_option past the list": func() error {
			return g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(2), nil)
		},
		"a negative pay_option": func() error {
			return g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(-1), nil)
		},
		"a pay_option on a Don't pay": func() error {
			return g.ResolvePayUnlessWithOption(c.ID, c.Chooser, false, intp(1), nil)
		},
		"cards on the damage branch": func() error {
			return g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(1), lands[:1])
		},
		"one land for the two-land branch": func() error {
			return g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(0), lands[:1])
		},
		"the same land twice": func() error {
			return g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(0), []uuid.UUID{lands[0], lands[0]})
		},
		"no cards for the sacrifice branch": func() error {
			return g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(0), nil)
		},
	}
	for name, answer := range cases {
		if err := answer(); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: %v, want ErrInvalidParam", name, err)
		}
		if got := onlyPendingChoice(t, g); got.ID != c.ID {
			t.Fatalf("%s: the prompt did not stay", name)
		}
	}
	if len(out.paidBy) != 0 || me.Life != 40 {
		t.Errorf("a refused answer paid (%v) or cost life (%d)", out.paidBy, me.Life)
	}
	for _, id := range lands {
		if !g.Battlefield.Contains(id) {
			t.Error("a refused answer sacrificed a land")
		}
	}
}

// A pay_option on a prompt that is not an either payment is refused,
// whatever the payment is: a mana prompt, a card pick, a names-nothing
// payment.
func TestAPayOptionOnAPlainPaymentIsRefused(t *testing.T) {
	for name, mutate := range map[string]func(*anyPlayerTestOffer){
		"mana":    func(p *anyPlayerTestOffer) { p.Cost = "{2}" },
		"damage":  withPayment(PayAction{Kind: PayActionDamage, Count: 5}, "5 damage"),
		"discard": withPayment(PayAction{Kind: PayActionDiscard, Count: 1}, "a card"),
	} {
		g := newFourPlayerActiveGame(t)
		var out offerOutcome
		queueOffer(t, g, g.Seats[2].ID, "{2}", &out, mutate)
		c := onlyPendingChoice(t, g)
		for _, apply := range []bool{true, false} {
			err := g.ResolvePayUnlessWithOption(c.ID, c.Chooser, apply, intp(0), nil)
			if !errors.Is(err, ErrInvalidParam) {
				t.Errorf("%s apply=%v: %v, want ErrInvalidParam", name, apply, err)
			}
		}
		if got := onlyPendingChoice(t, g); got.ID != c.ID || len(out.paidBy) != 0 {
			t.Errorf("%s: the prompt did not stay or something was paid", name)
		}
	}
}

// The sacrifice branch with fewer than two lands is not payable, and the
// damage branch always is; the offer is payable because one branch is.
func TestEitherPaymentBranchesArePayableOnTheirOwn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	landsFor(g, me, 1)
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "either", &out, wormsPayment())
	c := onlyPendingChoice(t, g)
	a := c.PayAction()
	if a == nil || a.Kind != PayActionEither || len(a.Either) != 2 {
		t.Fatalf("the prompt's payment is %+v, want an either of two", a)
	}
	var offer, sac, dmg bool
	g.WithWriteLock(func() {
		offer = g.PayActionCanPayForEffect(c.Chooser, a)
		sac = g.PayActionCanPayForEffect(c.Chooser, &a.Either[0])
		dmg = g.PayActionCanPayForEffect(c.Chooser, &a.Either[1])
	})
	if sac {
		t.Error("one land counted as paying for 'sacrifice two lands'")
	}
	if !dmg || !offer {
		t.Errorf("damage payable = %v, offer payable = %v; want both", dmg, offer)
	}
	// Naming the unpayable branch is a decline (CR 118.3), as an
	// unfundable "yes" always is: the offer moves on.
	if err := g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(0), nil); err != nil {
		t.Fatalf("naming an unpayable branch with no cards: %v, want a decline", err)
	}
	if len(out.paidBy) != 0 {
		t.Errorf("an unpayable branch paid: %v", out.paidBy)
	}
	if next := onlyPendingChoice(t, g); next.Chooser == c.Chooser {
		t.Error("the offer did not move on to the next player")
	}
}

// The sacrifice branch sacrifices exactly the lands named, and OnPaid
// runs once, with the payer.
func TestEitherPaymentSacrificeBranchPaysWithTheLandsNamed(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	lands := landsFor(g, me, 4)
	otherLand := landsFor(g, g.Seats[1], 1)[0]
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "either", &out, wormsPayment())
	c := onlyPendingChoice(t, g)

	// Another player's land is not one of the options.
	if err := g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(0), []uuid.UUID{lands[0], otherLand}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("another player's land: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(0), []uuid.UUID{lands[1], lands[3]}); err != nil {
		t.Fatalf("a legal pair: %v", err)
	}
	if g.Battlefield.Contains(lands[1]) || g.Battlefield.Contains(lands[3]) {
		t.Error("the named lands are still on the battlefield")
	}
	if !g.Battlefield.Contains(lands[0]) || !g.Battlefield.Contains(lands[2]) || !g.Battlefield.Contains(otherLand) {
		t.Error("a land that was not named was sacrificed")
	}
	if me.Life != 40 {
		t.Errorf("the sacrifice branch cost %d life", 40-me.Life)
	}
	if !sameIDs(out.paidBy, []uuid.UUID{me.ID}) || out.nobody != 0 {
		t.Errorf("paid by %v, nobody %d; want the active player and never", out.paidBy, out.nobody)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left open after the first payer", len(g.PendingChoices))
	}
}

// The damage branch is dealt by the offer's source and pays at once, with
// no lands named and none sacrificed.
func TestEitherPaymentDamageBranchDealsTheDamageAndPays(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	lands := landsFor(g, me, 3)
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "either", &out, wormsPayment())
	c := onlyPendingChoice(t, g)
	source := c.Source
	if err := g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(1), nil); err != nil {
		t.Fatal(err)
	}
	if me.Life != 35 {
		t.Errorf("life %d, want 40 - 5", me.Life)
	}
	for _, id := range lands {
		if !g.Battlefield.Contains(id) {
			t.Error("the damage branch sacrificed a land")
		}
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
	if !sameIDs(out.paidBy, []uuid.UUID{me.ID}) || out.nobody != 0 {
		t.Errorf("paid by %v, nobody %d; want the active player and never", out.paidBy, out.nobody)
	}
}

// CR 118.12: choosing damage that is then fully prevented still counts as
// paid. "If a player does either" asks whether they chose and paid it,
// not whether any damage landed.
func TestEitherPaymentDamageThatIsPreventedStillPays(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "either", &out, wormsPayment())
	g.WithWriteLock(func() {
		if !g.PreventDamageFromSourceThisTurnForEffect(DamageShield{Controller: me.ID, ProtectPlayer: me.ID, Label: "Fog"}) {
			t.Fatal("no shield registered")
		}
	})
	c := onlyPendingChoice(t, g)
	if err := g.ResolvePayUnlessWithOption(c.ID, c.Chooser, true, intp(1), nil); err != nil {
		t.Fatal(err)
	}
	if me.Life != 40 {
		t.Fatalf("life %d, want the damage prevented", me.Life)
	}
	if !sameIDs(out.paidBy, []uuid.UUID{me.ID}) {
		t.Errorf("paid by %v: prevented damage must still count as paid", out.paidBy)
	}
	if out.nobody != 0 {
		t.Error("OnNobody ran after a payment")
	}
}

// Everyone declining runs OnNobody once, and a decline never names a
// branch.
func TestEitherPaymentNobodyPaysRunsOnNobody(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "either", &out, wormsPayment())
	for i := 0; i < 4; i++ {
		c := onlyPendingChoice(t, g)
		if err := g.ResolvePayUnlessWithOption(c.ID, c.Chooser, false, nil, nil); err != nil {
			t.Fatalf("decline %d: %v", i, err)
		}
	}
	if out.nobody != 1 || len(out.paidBy) != 0 {
		t.Errorf("nobody %d, paid %v; want one OnNobody and no payer", out.nobody, out.paidBy)
	}
}

// An either with one branch, or one nested inside another, is a
// programming error: surfaced once, with the consequence still running.
func TestAMalformedEitherIsSurfacedAndTheConsequenceRuns(t *testing.T) {
	for name, a := range map[string]PayAction{
		"one branch": {Kind: PayActionEither, Either: []PayAction{{Kind: PayActionDamage, Count: 5}}},
		"nested": {Kind: PayActionEither, Either: []PayAction{
			{Kind: PayActionDamage, Count: 5},
			{Kind: PayActionEither, Either: []PayAction{{Kind: PayActionDamage, Count: 1}, {Kind: PayActionLife, Count: 1}}},
		}},
		"empty": {Kind: PayActionEither},
	} {
		g := newFourPlayerActiveGame(t)
		var out offerOutcome
		queueOffer(t, g, g.Seats[2].ID, "either", &out, withPayment(a, "either"))
		if len(g.PendingChoices) != 0 {
			t.Errorf("%s: a prompt was queued for a malformed either", name)
		}
		errored := false
		for _, ev := range g.Events {
			if ev.Kind == EventEffectError {
				errored = true
			}
		}
		if !errored {
			t.Errorf("%s: no EventEffectError", name)
		}
		if out.nobody != 1 {
			t.Errorf("%s: OnNobody ran %d times, want 1", name, out.nobody)
		}
	}
}
