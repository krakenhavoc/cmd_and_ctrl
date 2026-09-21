package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// attack_tax_test.go — ADR 0080 / #1063: CR 508.1a's cost to attack.
//
// Every test here is about a RULE rather than a card; Propaganda,
// Ghostly Prison and Sphere of Safety prove the same rules from the
// catalog side in cards/effects.

const (
	propagandaTestOracle    = "test-propaganda"
	ghostlyPrisonTestOracle = "test-ghostly-prison"
	sphereTestOracle        = "test-sphere-of-safety"
	taxForestTestOracle     = "test-tax-forest"
)

// withCatalogAttackTaxes installs a CatalogAttackTaxes shim for the
// duration of a test, the way withCatalogManaTriggers does for CR
// 605.1b triggers.
func withCatalogAttackTaxes(t *testing.T, fn func(key string) []AttackTax) {
	t.Helper()
	prev := CatalogAttackTaxes
	CatalogAttackTaxes = fn
	t.Cleanup(func() { CatalogAttackTaxes = prev })
}

// flatTax is Propaganda's clause as the engine reads it: one price,
// every attacking creature, the controller only.
func flatTax(mana string) AttackTax {
	return AttackTax{
		Label:    "Creatures can't attack you unless their controller pays " + mana + " for each of them.",
		ManaCost: func(AttackTaxQuery) string { return mana },
	}
}

// pushTaxEnchantment drops a permanent carrying `oracle` under
// `owner`. What it taxes is whatever the test's catalog shim says.
func pushTaxEnchantment(g *Game, owner *Player, name, oracle string) uuid.UUID {
	return pushBattlefieldForTest(g, owner.ID, name, "Enchantment", oracle)
}

// taxedTable is the shared setup: a two-seat game parked on declare
// attackers, one 2/2 under seat 0, and a shim serving `taxes` for
// every oracle ID the test seeded.
func taxedTable(t *testing.T, taxes map[string][]AttackTax) (*Game, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	withCatalogAttackTaxes(t, func(key string) []AttackTax { return taxes[key] })
	bear := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	return g, bear
}

// addMana fills a seat's pool with n colourless, which pays any
// generic cost and nothing coloured.
func addMana(p *Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(ManaToken{Color: "C"})
	}
}

// --- the declaration is refused unpaid, accepted paid --------------

func TestAttackUnderATaxIsRefusedWhenItIsNotPaid(t *testing.T) {
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
	})
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)

	err := g.DeclareAttacker(bear, g.Seats[1].ID)
	if !errors.Is(err, ErrAttackTaxUnpaid) {
		t.Fatalf("DeclareAttacker with an empty pool: %v, want ErrAttackTaxUnpaid", err)
	}
	var te *AttackTaxUnpaidError
	if !errors.As(err, &te) || te.Cost != "{2}" {
		t.Errorf("the error does not carry the price: %#v", te)
	}
	// Nothing happened: CR 508.1a is paid before the CR 508.1f tap,
	// so a refused declaration leaves the creature untapped and
	// undeclared.
	c := findCard(g, bear)
	if c == nil || c.Tapped || c.AttackingTarget != uuid.Nil {
		t.Errorf("a refused declaration staged something: tapped=%v target=%v", c.Tapped, c.AttackingTarget)
	}
}

func TestAttackUnderATaxIsAcceptedWhenItIsPaid(t *testing.T) {
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
	})
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	addMana(g.Seats[0], 2)

	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker with {2} in the pool: %v", err)
	}
	if c := findCard(g, bear); c == nil || c.AttackingTarget != g.Seats[1].ID {
		t.Fatal("the paid declaration did not stage the attacker")
	}
	if n := len(g.Seats[0].ManaPool); n != 0 {
		t.Errorf("mana left in the pool: %d, want 0 — the tax was charged", n)
	}
}

// --- all or nothing (CR 508.1a) ------------------------------------

func TestABulkDeclarationRefusesTheWholeSwingWhenTheTaxIsOnlyPartlyAffordable(t *testing.T) {
	g, first := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
	})
	second := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	// Enough for ONE attacker, not two.
	addMana(g.Seats[0], 2)

	declared, err := g.DeclareAttackers([]AttackDeclaration{
		{Attacker: first, Target: g.Seats[1].ID},
		{Attacker: second, Target: g.Seats[1].ID},
	})
	if !errors.Is(err, ErrAttackTaxUnpaid) {
		t.Fatalf("a half-payable swing: %v, want ErrAttackTaxUnpaid", err)
	}
	if len(declared) != 0 {
		t.Errorf("declared %d attackers, want 0 — the declaration is all or nothing", len(declared))
	}
	// And nothing was spent, so the player can still submit the
	// smaller declaration that is the real answer.
	if n := len(g.Seats[0].ManaPool); n != 2 {
		t.Errorf("mana pool = %d, want 2 — a refused declaration charges nothing", n)
	}
	if _, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: first, Target: g.Seats[1].ID}}); err != nil {
		t.Fatalf("the smaller declaration the player can afford: %v", err)
	}
}

// --- two taxes stack -----------------------------------------------

func TestTwoAttackTaxesOnOneDefenderStack(t *testing.T) {
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle:    {flatTax("{2}")},
		ghostlyPrisonTestOracle: {flatTax("{2}")},
	})
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	pushTaxEnchantment(g, g.Seats[1], "Ghostly Prison", ghostlyPrisonTestOracle)

	price := g.PriceAttackDeclaration([]AttackDeclaration{{Attacker: bear, Target: g.Seats[1].ID}})
	if price.Total.Generic != 4 {
		t.Errorf("one attacker into two {2} taxes = %d generic, want 4 (cost %q)", price.Total.Generic, price.Cost)
	}
	if len(price.Lines) != 2 {
		t.Errorf("breakdown has %d lines, want one per enchantment", len(price.Lines))
	}

	addMana(g.Seats[0], 3)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); !errors.Is(err, ErrAttackTaxUnpaid) {
		t.Fatalf("three mana against a stacked {4}: %v, want ErrAttackTaxUnpaid", err)
	}
	addMana(g.Seats[0], 1)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatalf("four mana against a stacked {4}: %v", err)
	}
}

func TestTheTaxScalesWithTheNumberOfAttackingCreatures(t *testing.T) {
	g, first := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
	})
	second := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	third := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)

	price := g.PriceAttackDeclaration([]AttackDeclaration{
		{Attacker: first, Target: g.Seats[1].ID},
		{Attacker: second, Target: g.Seats[1].ID},
		{Attacker: third, Target: g.Seats[1].ID},
	})
	// "{2} for each creature they control that's attacking you".
	if price.Cost != "{2}{2}{2}" || price.Total.Generic != 6 {
		t.Errorf("three attackers = %q / %d generic, want {2}{2}{2} / 6", price.Cost, price.Total.Generic)
	}
}

// --- a tax on one defender does not price attacks on another -------

func TestATaxOnOneDefenderDoesNotPriceAttacksOnAnother(t *testing.T) {
	withCatalogAttackTaxes(t, func(key string) []AttackTax {
		if key == propagandaTestOracle {
			return []AttackTax{flatTax("{2}")}
		}
		return nil
	})
	// Seat 0 attacks; seat 1 has the Propaganda; seat 2 does not.
	g3 := newActiveGameWithSeats(t, 3)
	bear := pushKeywordCreature(t, g3, g3.Seats[0], 2, 2)
	pushTaxEnchantment(g3, g3.Seats[1], "Propaganda", propagandaTestOracle)
	advanceIntoStep(t, g3, StepDeclareAttackers)

	taxed := g3.PriceAttackDeclaration([]AttackDeclaration{{Attacker: bear, Target: g3.Seats[1].ID}})
	if taxed.Total.Generic != 2 {
		t.Errorf("attacking the Propaganda seat costs %d, want 2", taxed.Total.Generic)
	}
	free := g3.PriceAttackDeclaration([]AttackDeclaration{{Attacker: bear, Target: g3.Seats[2].ID}})
	if !free.IsFree() {
		t.Errorf("attacking the OTHER seat costs %q, want nothing — the tax protects its controller only", free.Cost)
	}
	// And the engine agrees with its own pricer.
	if err := g3.DeclareAttacker(bear, g3.Seats[2].ID); err != nil {
		t.Fatalf("attacking the untaxed seat with an empty pool: %v", err)
	}
}

// --- the scope line: "you", vs "you or planeswalkers you control" ---

func TestAPlayerOnlyTaxDoesNotPriceAnAttackOnAPlaneswalker(t *testing.T) {
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
		sphereTestOracle: {func() AttackTax {
			t := flatTax("{2}")
			t.Scope = AttackTaxOnPlayerOrPlaneswalkers
			return t
		}()},
	})
	walker := pushBattlefieldForTest(g, g.Seats[1].ID, "Test Walker", "Legendary Planeswalker — Test", "test-walker")

	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	if p := g.PriceAttackDeclaration([]AttackDeclaration{{Attacker: bear, Target: walker}}); !p.IsFree() {
		t.Errorf(`"can't attack you" priced an attack on a planeswalker at %q, want free`, p.Cost)
	}

	pushTaxEnchantment(g, g.Seats[1], "Sphere of Safety", sphereTestOracle)
	if p := g.PriceAttackDeclaration([]AttackDeclaration{{Attacker: bear, Target: walker}}); p.Total.Generic != 2 {
		t.Errorf(`"you or planeswalkers you control" priced it at %d, want 2`, p.Total.Generic)
	}
}

// --- the tapper plans the tax, and mana triggers fire for it -------

func TestTheAutoTapperPlansTheAttackTax(t *testing.T) {
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
	})
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	lands := []uuid.UUID{
		pushBattlefieldForTest(g, g.Seats[0].ID, "Forest", "Basic Land — Forest", taxForestTestOracle),
		pushBattlefieldForTest(g, g.Seats[0].ID, "Forest", "Basic Land — Forest", taxForestTestOracle),
	}

	// Without AutoTap the pool is empty and the declaration is
	// refused, untapped lands or not: the zero value of the params is
	// the strict posture.
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); !errors.Is(err, ErrAttackTaxUnpaid) {
		t.Fatalf("no auto-tap, empty pool: %v, want ErrAttackTaxUnpaid", err)
	}
	if err := g.DeclareAttackerWith(bear, g.Seats[1].ID, DeclareAttackersParams{AutoTap: true}); err != nil {
		t.Fatalf("with auto-tap and two lands: %v", err)
	}
	for _, id := range lands {
		if c := findCard(g, id); c == nil || !c.Tapped {
			t.Errorf("land %s was not tapped for the attack tax", id)
		}
	}
}

func TestManaTriggersFireForTheTapsThatPayAnAttackTax(t *testing.T) {
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
	})
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	forest := pushBattlefieldForTest(g, g.Seats[0].ID, "Forest", "Basic Land — Forest", taxForestTestOracle)
	pushBattlefieldForTest(g, g.Seats[0].ID, "Forest", "Basic Land — Forest", taxForestTestOracle)
	pushAuraOn(g, g.Seats[0], "Wild Growth", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}")}
		}
		return nil
	})

	// The tapper plans two Forests for the {2} — it prices a land at
	// what the land itself makes and does not model triggered mana,
	// which is ADR 0074's line and not this test's subject. What IS
	// this test's subject: the declaration taps through
	// materializePlanLocked, so Wild Growth's CR 605.1b trigger fires
	// at the production site and the THIRD mana it makes is left
	// floating after the tax is spent.
	if err := g.DeclareAttackerWith(bear, g.Seats[1].ID, DeclareAttackersParams{AutoTap: true}); err != nil {
		t.Fatalf("two Forests against a {2} tax: %v", err)
	}
	if c := findCard(g, forest); c == nil || !c.Tapped {
		t.Error("the enchanted Forest was not tapped")
	}
	if n := len(g.Seats[0].ManaPool); n != 1 {
		t.Errorf("pool after the tax = %d, want 1 — Wild Growth's trigger did not fire for the taps", n)
	}
}

// --- the tax is a declaration-time question and nothing else -------

func TestATaxThatArrivesAfterTheDeclarationChargesNothing(t *testing.T) {
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
	})
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatalf("an untaxed declaration: %v", err)
	}
	// CR 508.1a is a turn-based action that has already happened.
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	advanceIntoStep(t, g, StepCombatDamage)
	if g.Seats[1].Life != 40-2 {
		t.Errorf("defender life = %d, want 38 — the attack was already declared", g.Seats[1].Life)
	}
}

func TestAnUnpayableTaxIsNotARestriction(t *testing.T) {
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
	})
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	// The creature is still perfectly able to attack — it has no
	// CantAttack bit, and AttackerEligible says so. What it lacks is
	// the mana. Keeping the two apart is what lets a card that
	// actually restricts attacking stay a restriction.
	if c := findCard(g, bear); c == nil || !CanAttack(c) || !AttackerEligible(c, g.Seats[0].ID) {
		t.Error("an attack tax marked the creature as unable to attack; it is a cost, not a prohibition")
	}
}

// --- a count-shaped tax --------------------------------------------

func TestACountedTaxIsPricedAtDeclarationTime(t *testing.T) {
	counted := AttackTax{
		Label: "pays {X} where X is the number of enchantments you control",
		ManaCost: func(q AttackTaxQuery) string {
			n := 0
			for i := range q.Game.Battlefield.Cards {
				if c := &q.Game.Battlefield.Cards[i]; c.Controller == q.Defender && c.IsEnchantment() {
					n++
				}
			}
			if n <= 0 {
				return ""
			}
			return "{" + itoa(n) + "}"
		},
	}
	g, bear := taxedTable(t, map[string][]AttackTax{sphereTestOracle: {counted}})
	pushTaxEnchantment(g, g.Seats[1], "Sphere of Safety", sphereTestOracle)

	if p := g.PriceAttackDeclaration([]AttackDeclaration{{Attacker: bear, Target: g.Seats[1].ID}}); p.Total.Generic != 1 {
		t.Errorf("a lone Sphere counts itself: %d, want 1", p.Total.Generic)
	}
	pushTaxEnchantment(g, g.Seats[1], "Other Enchantment", "test-other-enchantment")
	pushTaxEnchantment(g, g.Seats[1], "Third Enchantment", "test-third-enchantment")
	if p := g.PriceAttackDeclaration([]AttackDeclaration{{Attacker: bear, Target: g.Seats[1].ID}}); p.Total.Generic != 3 {
		t.Errorf("three enchantments: %d, want 3", p.Total.Generic)
	}
}

// --- the designation gate and CR 613.1f ability removal ------------

func TestAnAbilityRemovedPermanentTaxesNothing(t *testing.T) {
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
	})
	prop := pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	if p := g.PriceAttackDeclaration([]AttackDeclaration{{Attacker: bear, Target: g.Seats[1].ID}}); p.IsFree() {
		t.Fatal("setup: the tax should be charging before abilities are removed")
	}
	// CR 613.1f: the accessor reads CatalogAbilityKey, so a permanent
	// that has lost all abilities stops taxing — the same gate every
	// other catalog-contributed static goes through.
	g.WithWriteLock(func() {
		c := findCard(g, prop)
		eff := c.printedCharacteristic()
		eff.AbilitiesRemoved = true
		c.effective = &eff
	})
	if p := g.PriceAttackDeclaration([]AttackDeclaration{{Attacker: bear, Target: g.Seats[1].ID}}); !p.IsFree() {
		t.Errorf("a permanent with no abilities still taxed %q", p.Cost)
	}
}

// itoa is strconv.Itoa without the import, for the one test that
// renders a count.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
