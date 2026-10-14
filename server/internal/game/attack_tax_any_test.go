package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// attack_tax_any_test.go — #3456: the any-defender form of the timed
// attack tax (War Tax), priced by the same pricer as the rest.

func registerAnyTax(t *testing.T, g *Game, mana string) bool {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.AnyAttackTaxForEffect(uuid.Nil, mana, Duration{}, "Test — any tax")
	})
	return ok
}

func anyTaxPriceOf(g *Game, bear, target uuid.UUID) AttackTaxPrice {
	var p AttackTaxPrice
	g.WithWriteLock(func() {
		p = g.PriceAttackDeclarationForEffect([]AttackDeclaration{{Attacker: bear, Target: target}})
	})
	return p
}

func TestAnAnyDefenderTaxOfZeroWritesNothing(t *testing.T) {
	g, bear := taxedTable(t, nil)
	if registerAnyTax(t, g, "{0}") || registerAnyTax(t, g, "") {
		t.Fatal("a zero tax wrote a record")
	}
	if p := anyTaxPriceOf(g, bear, g.Seats[1].ID); !p.IsFree() {
		t.Errorf("price %q, want free", p.Cost)
	}
}

func TestAnAnyDefenderTaxChargesEveryAttackAndStacks(t *testing.T) {
	g, bear := taxedTable(t, nil)
	walker := pushBattlefieldForTest(g, g.Seats[1].ID, "Test Walker", "Legendary Planeswalker — Test", "test-walker")
	if !registerAnyTax(t, g, "{2}") {
		t.Fatal("not registered")
	}
	for name, target := range map[string]uuid.UUID{"player": g.Seats[1].ID, "planeswalker": walker} {
		if p := anyTaxPriceOf(g, bear, target); p.Total.Generic != 2 {
			t.Errorf("%s: price %q, want {2}", name, p.Cost)
		}
	}
	registerAnyTax(t, g, "{1}")
	if p := anyTaxPriceOf(g, bear, g.Seats[1].ID); p.Total.Generic != 3 || len(p.Lines) != 2 {
		t.Errorf("stacked: price %q with %d lines, want 3 from two", p.Cost, len(p.Lines))
	}
}

func TestAnAnyDefenderTaxIsRefusedUnpaidAndEndsAtEndOfTurn(t *testing.T) {
	g, bear := taxedTable(t, nil)
	registerAnyTax(t, g, "{2}")
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); !errors.Is(err, ErrAttackTaxUnpaid) {
		t.Fatalf("unpaid: %v, want ErrAttackTaxUnpaid", err)
	}
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	if p := anyTaxPriceOf(g, bear, g.Seats[0].ID); !p.IsFree() {
		t.Errorf("still taxed %q next turn", p.Cost)
	}
}

func TestAnAnyDefenderTaxRoundTripsThroughASnapshot(t *testing.T) {
	g, bear := taxedTable(t, nil)
	registerAnyTax(t, g, "{2}")
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if !snap.Restorable() {
		t.Fatal("not a restore point")
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if p := anyTaxPriceOf(restored, bear, restored.Seats[1].ID); p.Total.Generic != 2 {
		t.Errorf("restored price %q, want {2}", p.Cost)
	}
}
