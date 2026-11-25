package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cumulative_upkeep_payments_6189_test.go — #6189: the payments a
// cumulative upkeep prompt could not take (CR 702.24a: "you may pay
// [cost] for each age counter on it", and "if [cost] has choices
// associated with it, each choice is made separately for each age
// counter"), and mana that may pay nothing else (CR 106.6).

var cu6189Outcome struct{ paid, declined int }

var cu6189Branches = RegisterPayUnlessBranches(testEffectKeyPrefix+"cumulative-upkeep-6189",
	func(*Game, PayUnlessAnswer) error { cu6189Outcome.paid++; return nil },
	func(*Game, PayUnlessAnswer) error { cu6189Outcome.declined++; return nil })

// queueCU6189 asks `payer` for `a` (or `cost` when `a` is nil) as a
// cumulative upkeep's prompt would, and returns it.
func queueCU6189(t *testing.T, g *Game, payer, source uuid.UUID, a *PayAction, cost string, life int) *PendingChoice {
	t.Helper()
	cu6189Outcome.paid, cu6189Outcome.declined = 0, 0
	var err error
	resolvingTestRow(g, func() {
		err = g.QueuePayUnlessPromptKeyedForEffect(PayUnlessKeyedPrompt{
			Chooser: payer, Source: source, Cost: cost, Question: "upkeep?",
			Then: cu6189Branches, Action: a, LifeCost: life,
			SpendPurpose: SpendPurposeCumulativeUpkeep,
		})
	})
	if err != nil {
		t.Fatalf("queue: %v", err)
	}
	c := lastChoice(g)
	if c == nil || c.Kind != PendingChoicePayUnless {
		t.Fatalf("no pay-unless prompt: %+v", c)
	}
	return c
}

// Sheltering Ancient's payment: a +1/+1 counter on a creature an
// opponent controls, one per age counter, and the same creature may be
// named twice (CR 702.24a). Your own creature is never an option.
func TestCounterOnAnOpponentsCreaturePaymentTakesRepeats(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	mine := battlefieldCreatureForTest(g, a, "Mine", 2, 2)
	theirs := battlefieldCreatureForTest(g, b, "Theirs", 2, 2)
	pay := PayAction{
		Kind: PayActionCounterOnOpponentCreature, Count: 2, Repeats: true,
		CounterKind: CounterPlusOne, Of: PermanentQuery{Types: []string{"creature"}},
	}
	var opts []uuid.UUID
	var can bool
	g.WithWriteLock(func() {
		opts = g.PayActionOptionsForEffect(a.ID, &pay)
		can = g.PayActionCanPayForEffect(a.ID, &pay)
	})
	if len(opts) != 1 || opts[0] != theirs {
		t.Fatalf("options = %v, want only the opponent's creature", opts)
	}
	if !can {
		t.Error("one opponent's creature cannot pay two repeatable picks (CR 702.24a)")
	}
	c := queueCU6189(t, g, a.ID, uuid.Nil, &pay, "", 0)
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, []uuid.UUID{mine, theirs}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a creature of your own: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, []uuid.UUID{theirs, theirs}); err != nil {
		t.Fatalf("name the one creature twice: %v", err)
	}
	if cu6189Outcome.paid != 1 {
		t.Fatalf("outcome %+v, want the paid branch", cu6189Outcome)
	}
	if got := findBattlefieldCard(g, theirs).Counters[CounterPlusOne]; got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 (one per age counter)", got)
	}
}

// Aboroth's payment: N -1/-1 counters on the asking permanent itself,
// naming nothing. A permanent that can't have counters put on it cannot
// pay (CR 118.3).
func TestCountersOnTheSourcePaymentNamesNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a := g.Seats[0]
	aboroth := battlefieldCreatureForTest(g, a, "Aboroth", 9, 9)
	pay := PayAction{Kind: PayActionCountersOnSource, Count: 3, CounterKind: CounterMinusOne, OnSource: aboroth}
	if !pay.NamesNothing() {
		t.Fatal("counters on the source are a yes or no")
	}
	c := queueCU6189(t, g, a.ID, aboroth, &pay, "", 0)
	if err := g.ResolvePayUnless(c.ID, a.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if cu6189Outcome.paid != 1 {
		t.Fatalf("outcome %+v, want the paid branch", cu6189Outcome)
	}
	if got := findBattlefieldCard(g, aboroth).Counters[CounterMinusOne]; got != 3 {
		t.Errorf("-1/-1 counters = %d, want 3 (one per age counter)", got)
	}

	gone := PayAction{Kind: PayActionCountersOnSource, Count: 1, CounterKind: CounterMinusOne, OnSource: uuid.New()}
	var payable bool
	g.WithWriteLock(func() { payable = g.PayActionPayableForEffect(a.ID, &gone) })
	if payable {
		t.Error("a permanent that is not on the battlefield can pay counters on itself")
	}
}

// Wall of Shards' payment: an opponent the payer names gains the life,
// once per age counter, the same opponent twice if they like.
func TestOpponentGainsLifePaymentNamesPlayers(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	pay := PayAction{Kind: PayActionOpponentGainsLife, Count: 2, Amount: 1, PicksPlayers: true, Repeats: true}
	var opts []uuid.UUID
	g.WithWriteLock(func() { opts = g.PayActionOptionsForEffect(a.ID, &pay) })
	if len(opts) != 3 || containsID(opts, a.ID) {
		t.Fatalf("options = %v, want the payer's three opponents", opts)
	}
	start := b.Life
	c := queueCU6189(t, g, a.ID, uuid.Nil, &pay, "", 0)
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, []uuid.UUID{a.ID, b.ID}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("naming yourself: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, []uuid.UUID{b.ID, b.ID}); err != nil {
		t.Fatalf("name one opponent twice: %v", err)
	}
	if b.Life != start+2 {
		t.Errorf("the opponent's life %d → %d, want +2", start, b.Life)
	}
	if cu6189Outcome.paid != 1 {
		t.Errorf("outcome %+v, want the paid branch", cu6189Outcome)
	}
}

// Varchild's War-Riders' payment: a named opponent creates the token,
// so it enters under their control. An unregistered token key is a
// card-file bug, not a payment.
func TestOpponentCreatesTokenPaymentCreatesItUnderTheirControl(t *testing.T) {
	const key = "test/6189-survivor"
	RegisterPayActionToken(key, Card{Name: "Survivor", TypeLine: "Token Creature — Survivor", Power: 1, Toughness: 1, Colors: []string{"R"}})
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	pay := PayAction{Kind: PayActionOpponentCreatesToken, Count: 1, TokenKey: key, PicksPlayers: true, Repeats: true}
	c := queueCU6189(t, g, a.ID, uuid.Nil, &pay, "", 0)
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, []uuid.UUID{b.ID}); err != nil {
		t.Fatalf("pay: %v", err)
	}
	found := false
	for _, card := range g.Battlefield.Cards {
		if card.Name == "Survivor" {
			found = true
			if card.Controller != b.ID {
				t.Errorf("the token entered under %s, want the named opponent %s (CR 111.2)", card.Controller, b.ID)
			}
		}
	}
	if !found {
		t.Fatal("no token was created")
	}
	if cu6189Outcome.paid != 1 {
		t.Errorf("outcome %+v, want the paid branch", cu6189Outcome)
	}
}

// Herald of Leshrac's payment: the payer gains control of a land they
// don't control, one per age counter, each a different land.
func TestGainControlOfALandPaymentTakesALandYouDontControl(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	mine := landsFor(g, a, 1)[0]
	theirs := landsFor(g, b, 2)
	pay := PayAction{Kind: PayActionGainControlOfLand, Count: 1, Of: PermanentQuery{Types: []string{"land"}}}
	var opts []uuid.UUID
	g.WithWriteLock(func() { opts = g.PayActionOptionsForEffect(a.ID, &pay) })
	if len(opts) != 2 || containsID(opts, mine) {
		t.Fatalf("options = %v, want only the lands the payer does not control", opts)
	}
	c := queueCU6189(t, g, a.ID, uuid.Nil, &pay, "", 0)
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, []uuid.UUID{mine}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a land you control: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, a.ID, true, theirs[:1]); err != nil {
		t.Fatalf("pay: %v", err)
	}
	g.WithWriteLock(func() { g.recomputeLayersLocked() })
	if c := findBattlefieldCard(g, theirs[0]); c == nil || c.Controller != a.ID {
		t.Errorf("control of the land = %v, want the payer", c)
	}
	if c := findBattlefieldCard(g, theirs[1]); c == nil || c.Controller != b.ID {
		t.Error("the land that was not named changed hands")
	}
}

// Psychic Vortex draws, and Thought Lash exiles the top of the library:
// both name nothing, and the exile needs the cards (CR 118.3) while the
// draw is a payment even from an empty library.
func TestDrawAndExileLibraryTopPayments(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a := g.Seats[0]
	draw := PayAction{Kind: PayActionDraw, Count: 2}
	exile := PayAction{Kind: PayActionExileLibraryTop, Count: 2}
	hand, lib := a.Hand.Size(), a.Library.Size()
	c := queueCU6189(t, g, a.ID, uuid.Nil, &draw, "", 0)
	if err := g.ResolvePayUnless(c.ID, a.ID, true); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if a.Hand.Size() != hand+2 || a.Library.Size() != lib-2 {
		t.Errorf("hand %d → %d, library %d → %d; want two drawn", hand, a.Hand.Size(), lib, a.Library.Size())
	}
	lib = a.Library.Size()
	c = queueCU6189(t, g, a.ID, uuid.Nil, &exile, "", 0)
	if err := g.ResolvePayUnless(c.ID, a.ID, true); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if a.Library.Size() != lib-2 || g.Exile.Size() < 2 {
		t.Errorf("library %d → %d, exile %d; want the top two exiled", lib, a.Library.Size(), g.Exile.Size())
	}
	if cu6189Outcome.paid != 1 {
		t.Errorf("outcome %+v, want the paid branch", cu6189Outcome)
	}

	// An empty library: the draw is still a payment, the exile is not.
	a.Library.Cards = nil
	var canDraw, canExile bool
	g.WithWriteLock(func() {
		canDraw = g.PayActionPayableForEffect(a.ID, &draw)
		canExile = g.PayActionPayableForEffect(a.ID, &exile)
	})
	if !canDraw {
		t.Error("a draw from an empty library is not a payment (CR 104.3c says the player loses later)")
	}
	if canExile {
		t.Error("an empty library can pay an exile from the top (CR 118.3)")
	}
}

// Infernal Darkness' "Pay {B} and 1 life": one payment of both halves,
// and a payer who cannot pay the life has declined (CR 119.4).
func TestManaAndLifeCumulativeUpkeepPaysBoth(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a := g.Seats[0]
	a.ManaPool.AddMana(ManaToken{Color: "B"}, ManaToken{Color: "B"})
	start := a.Life
	c := queueCU6189(t, g, a.ID, uuid.Nil, nil, "{B}{B}", 2)
	if err := g.ResolvePayUnless(c.ID, a.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if cu6189Outcome.paid != 1 {
		t.Fatalf("outcome %+v, want the paid branch", cu6189Outcome)
	}
	if a.Life != start-2 {
		t.Errorf("life %d → %d, want 2 paid with the mana", start, a.Life)
	}

	a.ManaPool.AddMana(ManaToken{Color: "B"})
	a.Life = 1
	c = queueCU6189(t, g, a.ID, uuid.Nil, nil, "{B}", 2)
	if err := g.ResolvePayUnless(c.ID, a.ID, true); err != nil {
		t.Fatalf("an unpayable life half: %v", err)
	}
	if cu6189Outcome.declined != 1 {
		t.Errorf("outcome %+v, want the decline (CR 119.4)", cu6189Outcome)
	}
	if a.Life != 1 {
		t.Errorf("life = %d, want 1: nothing was paid", a.Life)
	}
}

// Snowfall's and Adarkar Unicorn's mana: "Spend this mana only to pay
// cumulative upkeep costs" (CR 106.6) pays a cumulative upkeep and
// nothing else.
func TestUpkeepOnlyManaPaysACumulativeUpkeepAndNothingElse(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a := g.Seats[0]
	restricted := ManaToken{Color: "U", Restrictions: []string{ManaRestrictCumulativeUpkeep}}

	a.ManaPool.AddMana(restricted)
	c := queueCU6189(t, g, a.ID, uuid.Nil, nil, "{U}", 0)
	if err := g.ResolvePayUnless(c.ID, a.ID, true); err != nil {
		t.Fatalf("pay a cumulative upkeep: %v", err)
	}
	if cu6189Outcome.paid != 1 {
		t.Fatalf("outcome %+v, want the paid branch: upkeep-only mana pays a cumulative upkeep", cu6189Outcome)
	}
	if len(a.ManaPool) != 0 {
		t.Errorf("mana pool = %v, want the restricted mana spent", a.ManaPool)
	}

	// The same mana on an ordinary pay-unless (a ward, a Rhystic tax):
	// no purpose, so the restriction refuses it (#352).
	a.ManaPool.AddMana(restricted)
	cu6189Outcome.paid, cu6189Outcome.declined = 0, 0
	var err error
	resolvingTestRow(g, func() {
		err = g.QueuePayUnlessPromptKeyedForEffect(PayUnlessKeyedPrompt{
			Chooser: a.ID, Cost: "{U}", Question: "ward", Then: cu6189Branches,
		})
	})
	if err != nil {
		t.Fatalf("queue the ward: %v", err)
	}
	ward := lastChoice(g)
	if err := g.ResolvePayUnless(ward.ID, a.ID, true); err != nil {
		t.Fatalf("answer the ward: %v", err)
	}
	if cu6189Outcome.declined != 1 {
		t.Errorf("outcome %+v, want the decline: upkeep-only mana cannot pay a ward", cu6189Outcome)
	}
	if len(a.ManaPool) != 1 {
		t.Errorf("mana pool = %v, want the restricted mana still there", a.ManaPool)
	}
}
