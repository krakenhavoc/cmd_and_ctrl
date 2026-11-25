package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// pay_unless_6187_test.go — #6187: three pay-unless payments (CR
// 118.12a: "[Do something] unless [a player does something else]" is
// "[a player may do something else]. If [that player doesn't], [do
// something]"). Balduvian Horde's "discard a card at random", Tainted
// Specter's "put a card from your hand on top of your library" and
// Phyrexian Dreadnought's "sacrifice any number of creatures with total
// power 12 or greater". The card side is effects.RandomDiscardPayment,
// HandToLibraryTopPayment and SacrificeTotalPowerPayment.

var payUnless6187Outcome struct{ paid, declined int }

var payUnless6187Branches = RegisterPayUnlessBranches(testEffectKeyPrefix+"pay-unless-6187",
	func(*Game, PayUnlessAnswer) error { payUnless6187Outcome.paid++; return nil },
	func(*Game, PayUnlessAnswer) error { payUnless6187Outcome.declined++; return nil })

// queue6187 asks `payer` for `a` as a resolving ability's pay-unless and
// returns the prompt.
func queue6187(t *testing.T, g *Game, payer uuid.UUID, a PayAction) *PendingChoice {
	t.Helper()
	payUnless6187Outcome.paid, payUnless6187Outcome.declined = 0, 0
	var err error
	resolvingTestRow(g, func() {
		err = g.QueuePayUnlessPromptKeyedForEffect(PayUnlessKeyedPrompt{
			Chooser: payer, Cost: PayActionPhrase(a), Question: "pay?", Then: payUnless6187Branches, Action: &a,
		})
	})
	if err != nil {
		t.Fatalf("queue: %v", err)
	}
	c := lastChoice(g)
	if c == nil || c.Kind != PendingChoicePayUnless {
		t.Fatalf("no pay-unless prompt queued: %+v", c)
	}
	return c
}

func handCardsFor(p *Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		c := NewCard("Card", p.ID)
		c.TypeLine = "Sorcery"
		p.Hand.PushTop(c)
		out = append(out, c.InstanceID)
	}
	return out
}

// A random discard names nothing: "yes" discards one card the engine
// picks from the hand, as a cost's discard (CR 701.9b), and the paid
// branch runs. An empty hand cannot pay, so its "yes" is the decline
// (CR 118.3).
func TestRandomDiscardPaymentDiscardsACardTheEnginePicks(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a := g.Seats[0]
	a.Hand.Cards = nil
	hand := handCardsFor(a, 3)
	pay := PayAction{Kind: PayActionDiscardRandom, Count: 1}
	if !pay.NamesNothing() {
		t.Fatal("a random discard is a yes or no: the payer names nothing")
	}
	c := queue6187(t, g, a.ID, pay)
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, hand[:1]); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("naming the card to discard: %v, want ErrInvalidParam (the engine picks, CR 701.9b)", err)
	}
	if err := g.ResolvePayUnless(c.ID, a.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if payUnless6187Outcome.paid != 1 || payUnless6187Outcome.declined != 0 {
		t.Fatalf("outcome %+v, want the paid branch", payUnless6187Outcome)
	}
	if a.Hand.Size() != 2 || a.Graveyard.Size() != 1 || !containsID(hand, a.Graveyard.Cards[0].InstanceID) {
		t.Errorf("hand %d, graveyard %d: want one of the three cards discarded", a.Hand.Size(), a.Graveyard.Size())
	}
	var cause DiscardCause
	for _, ev := range g.Events {
		if ev.Kind == EventDiscardCard && ev.Actor == a.ID {
			cause = ev.DiscardCause
		}
	}
	if cause != DiscardCauseCost {
		t.Errorf("discard cause %q, want a cost's (CR 118.12)", cause)
	}

	a.Hand.Cards = nil
	c = queue6187(t, g, a.ID, pay)
	var payable bool
	g.WithWriteLock(func() { payable = g.PayActionCanPayForEffect(a.ID, &pay) })
	if payable {
		t.Error("an empty hand can pay a random discard")
	}
	if err := g.ResolvePayUnless(c.ID, a.ID, true); err != nil {
		t.Fatalf("pay with an empty hand: %v", err)
	}
	if payUnless6187Outcome.declined != 1 || payUnless6187Outcome.paid != 0 {
		t.Errorf("outcome %+v, want the decline (CR 118.3)", payUnless6187Outcome)
	}
}

// Putting a hand card on top of the library is a card pick from the
// payer's hand, and the card goes on top. It is not a discard.
func TestHandToLibraryTopPaymentPutsTheNamedCardOnTop(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	a.Hand.Cards = nil
	hand := handCardsFor(a, 2)
	theirs := handCardsFor(b, 1)
	pay := PayAction{Kind: PayActionHandToLibraryTop, Count: 1}
	c := queue6187(t, g, a.ID, pay)
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, theirs); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("another player's hand card: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, hand); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("two cards for one: %v, want ErrInvalidParam", err)
	}
	discards := countEvents(g, EventDiscardCard)
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, hand[1:]); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if payUnless6187Outcome.paid != 1 {
		t.Fatalf("outcome %+v, want the paid branch", payUnless6187Outcome)
	}
	if top := a.Library.Cards[len(a.Library.Cards)-1]; top.InstanceID != hand[1] {
		t.Errorf("library top is %s, want the named card", top.Name)
	}
	if a.Hand.Contains(hand[1]) || !a.Hand.Contains(hand[0]) {
		t.Error("the wrong card left the hand")
	}
	if top := a.Library.Cards[len(a.Library.Cards)-1]; !top.KnownBy[a.ID] {
		t.Error("the payer no longer knows the card they put on top")
	}
	if countEvents(g, EventDiscardCard) != discards {
		t.Error("putting a card on top was announced as a discard")
	}

	a.Hand.Cards = nil
	c = queue6187(t, g, a.ID, pay)
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, nil); err != nil {
		t.Fatalf("pay with an empty hand: %v", err)
	}
	if payUnless6187Outcome.declined != 1 {
		t.Errorf("outcome %+v, want the decline (CR 118.3)", payUnless6187Outcome)
	}
}

// Phyrexian Dreadnought: any number of creatures, at least one, whose
// total power is 12 or more. 11 is refused, 12 pays, and a creature
// another player controls is never an option.
func TestSacrificeTotalPowerPaymentNeedsTheTotal(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	seven := battlefieldCreatureForTest(g, a, "Seven", 7, 7)
	four := battlefieldCreatureForTest(g, a, "Four", 4, 4)
	one := battlefieldCreatureForTest(g, a, "One", 1, 1)
	theirs := battlefieldCreatureForTest(g, b, "Their Giant", 12, 12)
	pay := PayAction{Kind: PayActionSacrificeTotalPower, TotalPower: 12, Of: PermanentQuery{Types: []string{"creature"}}}
	if pay.NamesNothing() {
		t.Fatal("a total-power sacrifice names the creatures")
	}
	c := queue6187(t, g, a.ID, pay)
	for name, picked := range map[string][]uuid.UUID{
		"total power 11":            {seven, four},
		"no creatures":              nil,
		"another player's creature": {theirs},
		"the same creature twice":   {seven, seven, four},
	} {
		if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, picked); name != "no creatures" && !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: %v, want ErrInvalidParam", name, err)
		}
		if name == "no creatures" {
			// "Pay" naming nothing while the payment can be made is
			// refused and the prompt stays.
			if lastChoice(g) == nil || lastChoice(g).ID != c.ID {
				t.Fatal("an empty answer took the prompt")
			}
		}
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, []uuid.UUID{seven, four, one}); err != nil {
		t.Fatalf("pay with total power 12: %v", err)
	}
	if payUnless6187Outcome.paid != 1 {
		t.Fatalf("outcome %+v, want the paid branch", payUnless6187Outcome)
	}
	for _, id := range []uuid.UUID{seven, four, one} {
		if g.Battlefield.Contains(id) {
			t.Error("a named creature was not sacrificed")
		}
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("another player's creature left")
	}

	// With 11 power on the board the payment cannot be made (CR 118.3).
	battlefieldCreatureForTest(g, a, "Eleven", 11, 11)
	var payable bool
	g.WithWriteLock(func() { payable = g.PayActionCanPayForEffect(a.ID, &pay) })
	if payable {
		t.Error("11 power can pay a total of 12")
	}
	battlefieldCreatureForTest(g, a, "One More", 1, 1)
	g.WithWriteLock(func() { payable = g.PayActionCanPayForEffect(a.ID, &pay) })
	if !payable {
		t.Error("12 power across two creatures cannot pay a total of 12")
	}
}
