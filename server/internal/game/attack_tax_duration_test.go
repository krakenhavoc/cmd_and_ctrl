package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// attack_tax_duration_test.go — #3398: an attack tax a resolved effect
// made for a duration (ModAttackTax), priced by the same pricer as the
// battlefield statics.

func registerTimedTax(t *testing.T, g *Game, defender uuid.UUID, mana string) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.AttackTaxForEffect(uuid.Nil, defender, mana, g.UntilYourNextTurnDuration(defender), "Test — timed tax")
	})
	if !ok {
		t.Fatal("AttackTaxForEffect registered nothing")
	}
}

func TestATimedAttackTaxIsRefusedUnpaidAndAcceptedPaid(t *testing.T) {
	g, bear := taxedTable(t, nil)
	registerTimedTax(t, g, g.Seats[1].ID, "{2}")

	if err := g.DeclareAttacker(bear, g.Seats[1].ID); !errors.Is(err, ErrAttackTaxUnpaid) {
		t.Fatalf("unpaid declaration: %v, want ErrAttackTaxUnpaid", err)
	}
	addMana(g.Seats[0], 2)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatalf("paid declaration: %v", err)
	}
	if n := len(g.Seats[0].ManaPool); n != 0 {
		t.Errorf("mana left in the pool: %d, want 0", n)
	}
}

func TestATimedAttackTaxStacksWithAStaticOne(t *testing.T) {
	g, bear := taxedTable(t, map[string][]AttackTax{propagandaTestOracle: {flatTax("{2}")}})
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	registerTimedTax(t, g, g.Seats[1].ID, "{2}")

	var price AttackTaxPrice
	g.WithWriteLock(func() {
		price = g.PriceAttackDeclarationForEffect([]AttackDeclaration{{Attacker: bear, Target: g.Seats[1].ID}})
	})
	if price.Cost != "{2}{2}" || len(price.Lines) != 2 {
		t.Errorf("price = %q with %d lines, want {2}{2} from two", price.Cost, len(price.Lines))
	}
}

func TestATimedAttackTaxEndsWithItsDuration(t *testing.T) {
	g, bear := taxedTable(t, nil)
	registerTimedTax(t, g, g.Seats[1].ID, "{2}")
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	var price AttackTaxPrice
	g.WithWriteLock(func() {
		price = g.PriceAttackDeclarationForEffect([]AttackDeclaration{{Attacker: bear, Target: g.Seats[1].ID}})
	})
	if !price.IsFree() {
		t.Errorf("the tax still prices %q once its owner's turn began", price.Cost)
	}
}

func TestAnAttackTaxModWithoutACostPanics(t *testing.T) {
	for name, m := range map[string]Mod{
		"no cost":  {Kind: ModAttackTax, Player: uuid.New()},
		"bad cost": {Kind: ModAttackTax, Player: uuid.New(), Text: "{zz}"},
	} {
		if attackTaxModProblem(m) == "" {
			t.Errorf("%s: accepted", name)
		}
	}
	if p := attackTaxModProblem(AttackTaxMod(uuid.New(), "{2}")); p != "" {
		t.Errorf("a sound mod refused: %s", p)
	}
}
