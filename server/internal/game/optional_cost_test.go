package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// optional_cost_test.go — ADR 0073 (#664), the engine half. The card
// files prove the mechanic; this file proves the plumbing under it:
// the announcement's validation, the record that rides the stack
// item, the buyback route's two branches, and that both survive an
// undo and a snapshot.
//
// It stubs the catalog hooks directly rather than registering cards,
// which is what lets it assert on the shapes no printed card has —
// two payments of a one-payment cost, an index out of range, a
// bought-back spell countered before it resolves.

const (
	testKickerOracle  = "test-optional-kicker"
	testBuybackOracle = "test-optional-buyback"
)

// stubOptionalCosts wires CatalogOptionalCosts for one oracle ID and
// restores the previous hook at the end of the test.
func stubOptionalCosts(t *testing.T, oracleID string, costs []AdditionalCost) {
	t.Helper()
	prev := CatalogOptionalCosts
	CatalogOptionalCosts = func(id string) []AdditionalCost {
		if id == oracleID {
			return costs
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { CatalogOptionalCosts = prev })
}

// castOptional seeds a card into the active seat's hand and announces
// it with the given optional-cost claim.
func castOptional(t *testing.T, g *Game, name, typeLine, oracleID string, optional []int) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		ManaCost:   "{1}",
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != StepPrecombatMain && g.Turn.Step != StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id, g.CastSpell(active.ID, id, CastSpellParams{OptionalCosts: optional})
}

// TestOptionalCostAnnouncementIsValidated is CR 601.2b's half: the
// claim itself, before anything is priced or paid.
func TestOptionalCostAnnouncementIsValidated(t *testing.T) {
	kicker := AdditionalCost{Optional: true, Key: KickerKey, ManaCost: "{2}", Label: "Kicker {2}"}
	multi := AdditionalCost{Optional: true, Key: MultikickerKey, ManaCost: "{1}", Repeat: 3, Label: "Multikicker {1}"}

	for _, tc := range []struct {
		name    string
		costs   []AdditionalCost
		claim   []int
		wantErr bool
	}{
		{"declining is always legal", []AdditionalCost{kicker}, nil, false},
		{"paying once", []AdditionalCost{kicker}, []int{0}, false},
		{"paying a once-only cost twice", []AdditionalCost{kicker}, []int{0, 0}, true},
		{"an index the card does not offer", []AdditionalCost{kicker}, []int{1}, true},
		{"a negative index", []AdditionalCost{kicker}, []int{-1}, true},
		{"claiming on a card that offers none", nil, []int{0}, true},
		{"multikicker up to its cap", []AdditionalCost{multi}, []int{0, 0, 0}, false},
		{"multikicker past its cap", []AdditionalCost{multi}, []int{0, 0, 0, 0}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOptionalCostChoice(tc.costs, tc.claim)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateOptionalCostChoice(%v) = %v, wantErr %v", tc.claim, err, tc.wantErr)
			}
		})
	}
}

// TestKickedCastRecordsThePaymentOnTheStackItem is the record itself:
// the fact the resolution reads, normalised, and present only when
// the cost was actually claimed.
func TestKickedCastRecordsThePaymentOnTheStackItem(t *testing.T) {
	g := newRestorableGame(t)
	stubOptionalCosts(t, testKickerOracle, []AdditionalCost{
		{Optional: true, Key: MultikickerKey, ManaCost: "{1}", Repeat: 5, Label: "Multikicker {1}"},
	})

	id, err := castOptional(t, g, "Kicked Thing", "Sorcery", testKickerOracle, []int{0, 0})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	item := g.StackMeta[id]
	if item == nil {
		t.Fatalf("no stack item for the cast spell")
	}
	if got := len(item.Paid.OptionalCosts); got != 2 {
		t.Fatalf("Paid.OptionalCosts = %v, want two entries", item.Paid.OptionalCosts)
	}
	if n := item.Paid.OptionalCostTimes(0); n != 2 {
		t.Errorf("OptionalCostTimes(0) = %d, want 2", n)
	}
	if item.Paid.PaidOptionalCost(1) {
		t.Errorf("an index the card does not offer reads as paid")
	}
	if item.Paid.IsZero() {
		t.Errorf("a record carrying optional costs reports IsZero")
	}
}

func TestUnkickedCastRecordsNothing(t *testing.T) {
	g := newRestorableGame(t)
	stubOptionalCosts(t, testKickerOracle, []AdditionalCost{
		{Optional: true, Key: KickerKey, ManaCost: "{2}", Label: "Kicker {2}"},
	})

	id, err := castOptional(t, g, "Unkicked Thing", "Sorcery", testKickerOracle, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if got := g.StackMeta[id].Paid.OptionalCosts; len(got) != 0 {
		t.Errorf("an unkicked cast recorded %v", got)
	}
}

// TestKickerAddsItsManaToTheCost is the CR 601.2f half: the kicked
// cast owes more, and the unkicked one does not.
func TestKickerAddsItsManaToTheCost(t *testing.T) {
	base, err := ParseCost("{1}{R}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	costs := []AdditionalCost{
		{Optional: true, Key: KickerKey, ManaCost: "{2}{W}", Label: "Kicker {2}{W}"},
	}

	unkicked, err := AddOptionalCostMana(base, costs, nil)
	if err != nil {
		t.Fatalf("AddOptionalCostMana(nil): %v", err)
	}
	if unkicked.Generic != base.Generic || len(unkicked.Required) != len(base.Required) {
		t.Errorf("declining the kicker changed the cost: %+v", unkicked)
	}

	kicked, err := AddOptionalCostMana(base, costs, []int{0})
	if err != nil {
		t.Fatalf("AddOptionalCostMana([0]): %v", err)
	}
	if kicked.Generic != base.Generic+2 {
		t.Errorf("kicked generic = %d, want %d", kicked.Generic, base.Generic+2)
	}
	if len(kicked.Required) != len(base.Required)+1 {
		t.Errorf("kicked coloured slots = %d, want %d", len(kicked.Required), len(base.Required)+1)
	}

	// Multikicker paid three times charges three times.
	multi := []AdditionalCost{{Optional: true, Key: MultikickerKey, ManaCost: "{1}", Repeat: 9}}
	thrice, err := AddOptionalCostMana(base, multi, []int{0, 0, 0})
	if err != nil {
		t.Fatalf("AddOptionalCostMana([0,0,0]): %v", err)
	}
	if thrice.Generic != base.Generic+3 {
		t.Errorf("thrice-kicked generic = %d, want %d", thrice.Generic, base.Generic+3)
	}
}

// TestBuybackRoutesOnlyOnResolution is CR 702.27a's "as it resolves",
// stated as the difference between the two exits a spell has.
func TestBuybackRoutesOnlyOnResolution(t *testing.T) {
	buyback := []AdditionalCost{
		{Optional: true, Key: BuybackKey, ManaCost: "{3}", Label: "Buyback {3}"},
	}

	t.Run("resolving returns it to hand", func(t *testing.T) {
		g := newRestorableGame(t)
		stubOptionalCosts(t, testBuybackOracle, buyback)
		active := g.Seats[g.Turn.ActiveSeat]

		id, err := castOptional(t, g, "Bought Back", "Instant", testBuybackOracle, []int{0})
		if err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		g.WithWriteLock(func() {
			if err := g.resolveTopOfStackLocked(); err != nil {
				t.Fatalf("resolve: %v", err)
			}
		})
		if !active.Hand.Contains(id) {
			t.Errorf("a bought-back spell did not return to its owner's hand")
		}
		if active.Graveyard.Contains(id) {
			t.Errorf("a bought-back spell also reached the graveyard")
		}
	})

	t.Run("countered by game rules goes to the graveyard", func(t *testing.T) {
		g := newRestorableGame(t)
		stubOptionalCosts(t, testBuybackOracle, buyback)
		active := g.Seats[g.Turn.ActiveSeat]

		id, err := castOptional(t, g, "Bought Back", "Instant", testBuybackOracle, []int{0})
		if err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		// The fizzle exit, reached directly: CR 608.2b's "countered
		// by game rules" is the branch that passes resolved=false.
		g.WithWriteLock(func() {
			card := g.Stack.Cards[len(g.Stack.Cards)-1]
			item := g.StackMeta[id]
			delete(g.StackMeta, id)
			if err := g.routeStackCardToGraveyardLocked(card, item, false); err != nil {
				t.Fatalf("route: %v", err)
			}
		})
		if active.Hand.Contains(id) {
			t.Errorf("a bought-back spell that never resolved was returned to hand")
		}
		if !active.Graveyard.Contains(id) {
			t.Errorf("the spell did not reach the graveyard")
		}
	})

	t.Run("declining buyback goes to the graveyard", func(t *testing.T) {
		g := newRestorableGame(t)
		stubOptionalCosts(t, testBuybackOracle, buyback)
		active := g.Seats[g.Turn.ActiveSeat]

		id, err := castOptional(t, g, "Not Bought Back", "Instant", testBuybackOracle, nil)
		if err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		g.WithWriteLock(func() {
			if err := g.resolveTopOfStackLocked(); err != nil {
				t.Fatalf("resolve: %v", err)
			}
		})
		if !active.Graveyard.Contains(id) {
			t.Errorf("an unbought spell did not reach the graveyard")
		}
	})
}

// TestKickedCastSurvivesUndo is the clone half: the record rides the
// stack item, so restoring a snapshot taken while a kicked spell is
// on the stack brings the kick back with it — and a restore taken
// BEFORE the cast does not invent one.
func TestKickedCastSurvivesUndo(t *testing.T) {
	g := newRestorableGame(t)
	stubOptionalCosts(t, testKickerOracle, []AdditionalCost{
		{Optional: true, Key: MultikickerKey, ManaCost: "{1}", Repeat: 5, Label: "Multikicker {1}"},
	})

	id, err := castOptional(t, g, "Kicked Thing", "Sorcery", testKickerOracle, []int{0, 0})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	mid := g.Clone()

	// The clone is a DEEP copy: mutating the live record must not
	// reach it, or an undo would restore the state it was meant to
	// roll back.
	g.WithWriteLock(func() {
		g.StackMeta[id].Paid.OptionalCosts[0] = 99
		delete(g.StackMeta, id)
	})
	if got := mid.StackMeta[id].Paid.OptionalCosts[0]; got != 0 {
		t.Fatalf("the clone aliases the live record: index 0 became %d", got)
	}

	// And the undo brings the kick back with the item.
	g.RestoreFrom(mid)
	item := g.StackMeta[id]
	if item == nil {
		t.Fatalf("the undo lost the stack item")
	}
	if n := item.Paid.OptionalCostTimes(0); n != 2 {
		t.Errorf("after undo, OptionalCostTimes(0) = %d, want 2", n)
	}
}

// TestPaidOptionalCostsSurviveTheSnapshot is the deploy half: the
// permanent's carried record is a fact about a spell that has already
// left the stack, so nothing could rebuild it.
//
// The record moved under Card.Provenance in #719 — one per-entry "how
// was this spell cast" record rather than two fields with the same
// lifecycle. What it has to survive is unchanged.
func TestPaidOptionalCostsSurviveTheSnapshot(t *testing.T) {
	g := newRestorableGame(t)
	owner := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Kicked Permanent",
		TypeLine:   "Creature — Elemental",
		Owner:      owner.ID,
		Controller: owner.ID,
		Provenance: CastProvenance{OptionalCosts: []int{0, 0, 0}},
	})

	snap := g.CaptureSnapshot()
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := decoded.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	var got []int
	for _, c := range restored.Battlefield.Cards {
		if c.InstanceID == id {
			got = c.Provenance.OptionalCosts
		}
	}
	if len(got) != 3 {
		t.Fatalf("Provenance.OptionalCosts after a round trip = %v, want three entries", got)
	}
}

// TestPaidOptionalCostsClearOnTheWayOut is CR 400.7: a permanent that
// leaves and comes back is a new object, and it was not kicked.
func TestPaidOptionalCostsClearOnTheWayOut(t *testing.T) {
	g := newRestorableGame(t)
	owner := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Kicked Permanent",
		TypeLine:   "Creature — Elemental",
		Owner:      owner.ID,
		Controller: owner.ID,
		Provenance: CastProvenance{OptionalCosts: []int{0}},
	})

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	for _, c := range owner.Hand.Cards {
		if c.InstanceID == id && len(c.Provenance.OptionalCosts) != 0 {
			t.Errorf("the kicked record survived the permanent leaving: %v", c.Provenance.OptionalCosts)
		}
	}
}

// TestAdditionalCostPlanWalksTheFlatLists is the payment-plan half:
// one validator, two clauses, and a flat wire list walked in plan
// order.
func TestAdditionalCostPlanWalksTheFlatLists(t *testing.T) {
	mandatory := &AdditionalCost{DiscardCards: 1, Label: "Discard a card"}
	optional := []AdditionalCost{
		{Optional: true, Key: KickerKey, DiscardCards: 2, Label: "Kicker—Discard two cards"},
	}

	plan := castCostPayments(mandatory, optional, nil)
	if len(plan) != 1 || plan[0].index != -1 {
		t.Fatalf("unkicked plan = %+v, want the mandatory cost alone", plan)
	}
	plan = castCostPayments(mandatory, optional, []int{0})
	if len(plan) != 2 || plan[0].index != -1 || plan[1].index != 0 {
		t.Fatalf("kicked plan = %+v, want mandatory then optional 0", plan)
	}
	if got := plan[0].cost.DiscardCards + plan[1].cost.DiscardCards; got != 3 {
		t.Errorf("the plan demands %d discards, want 3", got)
	}

	// The record normalises whatever order the client sent.
	paid := paidWithOptionalCosts(PaidCost{}, castCostPayments(nil,
		[]AdditionalCost{{Optional: true, Key: "a"}, {Optional: true, Key: "b"}},
		[]int{1, 0, 1}))
	want := []int{0, 1, 1}
	if len(paid.OptionalCosts) != len(want) {
		t.Fatalf("record = %v, want %v", paid.OptionalCosts, want)
	}
	for i := range want {
		if paid.OptionalCosts[i] != want[i] {
			t.Fatalf("record = %v, want %v (ascending)", paid.OptionalCosts, want)
		}
	}
}

// TestCastRefusesAnUnofferedOptionalCost pins that the wire is kept
// honest: an announcement naming a cost the card does not offer is a
// rejected cast, not a silently ignored field.
func TestCastRefusesAnUnofferedOptionalCost(t *testing.T) {
	g := newRestorableGame(t)
	active := g.Seats[g.Turn.ActiveSeat]

	id, err := castOptional(t, g, "Plain Spell", "Sorcery", "test-no-optional-costs", []int{0})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("CastSpell = %v, want ErrInvalidParam", err)
	}
	if !active.Hand.Contains(id) {
		t.Errorf("a refused cast did not leave the card in hand")
	}
}
