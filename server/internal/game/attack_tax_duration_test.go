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

// #5869: the life-priced sibling (Sivitri, Dragon Master).

func TestAnAttackTaxLifeModNeedsAPlayerAndLife(t *testing.T) {
	for name, m := range map[string]Mod{
		"no player": {Kind: ModAttackTaxLife, Amount: 2},
		"no life":   {Kind: ModAttackTaxLife, Player: uuid.New()},
		"bad types": {Kind: ModAttackTaxLife, Player: uuid.New(), Amount: 2, Types: []string{"battle"}},
	} {
		if attackTaxModProblem(m) == "" {
			t.Errorf("%s: accepted", name)
		}
	}
	for _, scope := range []AttackTaxScope{AttackTaxOnPlayer, AttackTaxOnPlayerOrPlaneswalkers} {
		if p := attackTaxModProblem(AttackTaxLifeMod(uuid.New(), 2, scope)); p != "" {
			t.Errorf("a sound mod refused: %s", p)
		}
	}
}

// A static tax may name both halves: the mana is paid, then the life,
// and a seat with the mana but not the life pays neither (CR 119.4).
func TestAStaticTaxWithManaAndLifeIsAllOrNothing(t *testing.T) {
	mixed := AttackTax{
		Label:    "Test — mana and life",
		ManaCost: func(AttackTaxQuery) string { return "{1}" },
		LifeCost: func(AttackTaxQuery) int { return 3 },
	}
	g, bear := taxedTable(t, map[string][]AttackTax{propagandaTestOracle: {mixed}})
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)

	price := g.PriceAttackDeclaration([]AttackDeclaration{{Attacker: bear, Target: g.Seats[1].ID}})
	if price.Cost != "{1}" || price.Life != 3 || price.Describe() != "{1} and 3 life" {
		t.Fatalf("price = %q / %d life (%q), want {1} and 3 life", price.Cost, price.Life, price.Describe())
	}
	addMana(g.Seats[0], 1)
	g.Seats[0].Life = 2
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); !errors.Is(err, ErrAttackTaxUnpaid) {
		t.Fatalf("2 life against 3: %v, want ErrAttackTaxUnpaid", err)
	}
	if n := len(g.Seats[0].ManaPool); n != 1 || g.Seats[0].Life != 2 {
		t.Fatalf("a refused declaration spent something: pool %d, life %d", n, g.Seats[0].Life)
	}
	g.Seats[0].Life = 3
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatalf("paid declaration: %v", err)
	}
	if n := len(g.Seats[0].ManaPool); n != 0 || g.Seats[0].Life != 0 {
		t.Errorf("after paying: pool %d, life %d, want 0 and 0", n, g.Seats[0].Life)
	}
}

// S82 OLD batch 181: Forbidding Spirit's "you or planeswalkers you control".

func TestATimedAttackTaxCanDefendPlaneswalkersToo(t *testing.T) {
	g, bear := taxedTable(t, nil)
	defender := g.Seats[1].ID
	var ok bool
	g.WithWriteLock(func() {
		ok = g.AttackTaxPlaneswalkersForEffect(uuid.Nil, defender, "{2}", g.UntilYourNextTurnDuration(defender), "Test — walker tax")
	})
	if !ok {
		t.Fatal("AttackTaxPlaneswalkersForEffect registered nothing")
	}
	var onPlayer, onWalker []AttackTaxLine
	g.WithWriteLock(func() {
		onPlayer = g.durationAttackTaxLines(defender, AttackTargetPlayer, bear)
		onWalker = g.durationAttackTaxLines(defender, AttackTargetPlaneswalker, bear)
	})
	if len(onPlayer) != 1 || len(onWalker) != 1 {
		t.Errorf("lines on the player %d and on a planeswalker %d, want 1 and 1", len(onPlayer), len(onWalker))
	}

	// The plain mana tax still leaves planeswalkers alone.
	g2, bear2 := taxedTable(t, nil)
	registerTimedTax(t, g2, g2.Seats[1].ID, "{2}")
	g2.WithWriteLock(func() {
		onWalker = g2.durationAttackTaxLines(g2.Seats[1].ID, AttackTargetPlaneswalker, bear2)
	})
	if len(onWalker) != 0 {
		t.Errorf("the plain tax charged an attack on a planeswalker: %d lines", len(onWalker))
	}
}

func TestAnAttackTaxModNamesOnlyPlaneswalkerTypes(t *testing.T) {
	if p := attackTaxModProblem(AttackTaxPlaneswalkersMod(uuid.New(), "{2}")); p != "" {
		t.Errorf("a sound mod refused: %s", p)
	}
	bad := AttackTaxMod(uuid.New(), "{2}")
	bad.Types = []string{"battle"}
	if attackTaxModProblem(bad) == "" {
		t.Error("a mod naming battle was accepted")
	}
}
