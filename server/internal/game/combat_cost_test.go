package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// combat_cost_test.go — #6190: a NON-MANA cost to attack or block
// (CR 508.1h, 509.1d). Every test here is about a rule rather than a
// card; the cards (Exalted Dragon, Leviathan, Flooded Woodlands,
// Reclamation, Hollow Warrior, Floodtide Serpent) prove the same rules
// from the catalog side when their batches build them.

const (
	sacrificeTaxOracle = "test-sacrifice-to-attack"
	tapTaxOracle       = "test-tap-to-attack"
	returnTaxOracle    = "test-return-to-attack"
)

// landSpec is "a land you control", the clause Exalted Dragon prints.
func landSpec(n int) *TargetSpec {
	return &TargetSpec{
		Mode: "permanent", Label: "a land", Min: n, Max: n,
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsLand()
		},
	}
}

// islandSpec is Leviathan's "two Islands".
func islandSpec(n int) *TargetSpec {
	return &TargetSpec{
		Mode: "permanent", Label: "two Islands", Min: n, Max: n,
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsLand() && c.HasSubtype("Island")
		},
	}
}

// sacrificeToAttackTax is Exalted Dragon's clause as the engine reads
// it: a cost the ATTACKING creature's controller owes whatever they
// attack, bound to the permanent that prints it.
func sacrificeToAttackTax(spec *TargetSpec, label string) AttackTax {
	return AttackTax{
		Label:      "This creature can't attack unless you sacrifice " + label + ".",
		OnAttacker: true,
		AppliesTo: func(q AttackTaxQuery) bool {
			return q.Attacker.InstanceID == q.Source.InstanceID
		},
		NonMana: CombatCost{Sacrifice: spec, Label: label},
	}
}

// greenAttackersSacrificeALand is Flooded Woodlands: every green
// attacker, whoever controls it, costs its controller a land.
func greenAttackersSacrificeALand() AttackTax {
	return AttackTax{
		Label:      "Green creatures can't attack unless their controller sacrifices a land of their choice for each green creature they control that's attacking.",
		OnAttacker: true,
		AppliesTo: func(q AttackTaxQuery) bool {
			return q.Attacker.HasColor("G")
		},
		NonMana: CombatCost{Sacrifice: landSpec(1), Label: "a land"},
	}
}

// pushUnsickCreature puts a 2/2 of the given mana cost (and so colour)
// under `owner`, able to attack.
func pushUnsickCreature(g *Game, owner *Player, name, manaCost string) uuid.UUID {
	id := pushColoredCreature(g, owner, name, manaCost)
	if c := findCard(g, id); c != nil {
		c.SummonedThisTurn = false
	}
	return id
}

// pushTestLand puts an untapped land of `typeLine` under `owner`.
func pushTestLand(g *Game, owner *Player, name, typeLine string) uuid.UUID {
	return pushTypedTestCard(g, Card{
		Name: name, TypeLine: typeLine, Owner: owner.ID, Controller: owner.ID,
	})
}

// TestAttackCostSacrificeIsRefusedUnpaidAndPaidWhenNamed — CR 508.1h:
// the land is named on the declaration, and a declaration that names
// none is refused with nothing staged, tapped or sacrificed.
func TestAttackCostSacrificeIsRefusedUnpaidAndPaidWhenNamed(t *testing.T) {
	g, dragon := taxedTable(t, map[string][]AttackTax{
		sacrificeTaxOracle: {sacrificeToAttackTax(landSpec(1), "a land")},
	})
	g.WithWriteLock(func() { findBattlefieldCard(g, dragon).OracleID = sacrificeTaxOracle })
	land := pushTestLand(g, g.Seats[0], "Plains", "Basic Land — Plains")

	err := g.DeclareAttacker(dragon, g.Seats[1].ID)
	if !errors.Is(err, ErrCombatCostUnpaid) {
		t.Fatalf("DeclareAttacker naming no land = %v, want ErrCombatCostUnpaid", err)
	}
	var ce *CombatCostUnpaidError
	if !errors.As(err, &ce) || ce.Kind != CombatCostSacrifice || ce.Clause != "a land" {
		t.Errorf("the error does not carry the clause: %#v", ce)
	}
	if c := findCard(g, dragon); c.Tapped || c.AttackingTarget != uuid.Nil {
		t.Errorf("a refused declaration staged something: tapped=%v target=%v", c.Tapped, c.AttackingTarget)
	}
	if findBattlefieldCard(g, land) == nil {
		t.Fatal("a refused declaration sacrificed the land")
	}

	if err := g.DeclareAttackerWith(dragon, g.Seats[1].ID, DeclareAttackersParams{
		Pay: CombatCostPayment{SacrificeIDs: []uuid.UUID{land}},
	}); err != nil {
		t.Fatalf("DeclareAttacker naming the land: %v", err)
	}
	if c := findCard(g, dragon); c.AttackingTarget != g.Seats[1].ID {
		t.Fatal("the paid declaration did not stage the attacker")
	}
	if findBattlefieldCard(g, land) != nil {
		t.Error("the land was not sacrificed")
	}
	if g.Seats[0].Graveyard.Size() != 1 {
		t.Errorf("graveyard = %d cards, want the sacrificed land", g.Seats[0].Graveyard.Size())
	}
}

// TestAttackCostRefusesAPermanentTheClauseDoesNotName — the clause is
// matched by the component's own validator, so a creature does not pay
// "sacrifice a land".
func TestAttackCostRefusesAPermanentTheClauseDoesNotName(t *testing.T) {
	g, dragon := taxedTable(t, map[string][]AttackTax{
		sacrificeTaxOracle: {sacrificeToAttackTax(landSpec(1), "a land")},
	})
	g.WithWriteLock(func() { findBattlefieldCard(g, dragon).OracleID = sacrificeTaxOracle })
	other := pushKeywordCreature(t, g, g.Seats[0], 1, 1)
	pushTestLand(g, g.Seats[0], "Plains", "Basic Land — Plains")

	err := g.DeclareAttackerWith(dragon, g.Seats[1].ID, DeclareAttackersParams{
		Pay: CombatCostPayment{SacrificeIDs: []uuid.UUID{other}},
	})
	if !errors.Is(err, ErrCombatCostUnpaid) || !errors.Is(err, ErrIllegalTarget) {
		t.Fatalf("naming a creature for \"sacrifice a land\" = %v, want an illegal-choice refusal", err)
	}
	if findBattlefieldCard(g, other) == nil {
		t.Error("the refused declaration sacrificed the creature anyway")
	}
}

// TestAttackCostMatchesNamedPermanentsToTheDemands — two clauses in one
// declaration ("a land" and "two Islands") are three slots, and the
// matching is exact: two Islands and a Forest pay, one Island and two
// Forests do not.
func TestAttackCostMatchesNamedPermanentsToTheDemands(t *testing.T) {
	g, dragon := taxedTable(t, map[string][]AttackTax{
		sacrificeTaxOracle: {sacrificeToAttackTax(landSpec(1), "a land")},
		returnTaxOracle:    {sacrificeToAttackTax(islandSpec(2), "two Islands")},
	})
	leviathan := pushKeywordCreature(t, g, g.Seats[0], 5, 5)
	g.WithWriteLock(func() {
		findBattlefieldCard(g, dragon).OracleID = sacrificeTaxOracle
		findBattlefieldCard(g, leviathan).OracleID = returnTaxOracle
		findBattlefieldCard(g, leviathan).SummonedThisTurn = false
	})
	i1 := pushTestLand(g, g.Seats[0], "Island", "Basic Land — Island")
	i2 := pushTestLand(g, g.Seats[0], "Island", "Basic Land — Island")
	forest := pushTestLand(g, g.Seats[0], "Forest", "Basic Land — Forest")
	f2 := pushTestLand(g, g.Seats[0], "Forest", "Basic Land — Forest")

	// One Island and two Forests cannot cover "two Islands".
	_, err := g.DeclareAttackersWith([]AttackDeclaration{
		{Attacker: dragon, Target: g.Seats[1].ID},
		{Attacker: leviathan, Target: g.Seats[1].ID},
	}, DeclareAttackersParams{
		Pay: CombatCostPayment{SacrificeIDs: []uuid.UUID{i1, forest, f2}},
	})
	if !errors.Is(err, ErrCombatCostUnpaid) {
		t.Fatalf("one Island for \"two Islands\" = %v, want ErrCombatCostUnpaid", err)
	}
	// Two Islands and a Forest do: the Forest pays "a land".
	if _, err := g.DeclareAttackersWith([]AttackDeclaration{
		{Attacker: dragon, Target: g.Seats[1].ID},
		{Attacker: leviathan, Target: g.Seats[1].ID},
	}, DeclareAttackersParams{
		Pay: CombatCostPayment{SacrificeIDs: []uuid.UUID{i1, i2, forest}},
	}); err != nil {
		t.Fatalf("two Islands and a Forest: %v", err)
	}
	if findBattlefieldCard(g, f2) == nil {
		t.Error("a land nobody named was sacrificed")
	}
}

// TestAttackCostIsPerTaxedAttacker — Flooded Woodlands' "for each green
// creature they control that's attacking": two green attackers owe two
// lands, and a nongreen one owes none.
func TestAttackCostIsPerTaxedAttacker(t *testing.T) {
	g := newActiveGame(t)
	withCatalogAttackTaxes(t, func(key string) []AttackTax {
		if key == sacrificeTaxOracle {
			return []AttackTax{greenAttackersSacrificeALand()}
		}
		return nil
	})
	me, them := g.Seats[0], g.Seats[1]
	// The enchantment is the OPPONENT's: the clause binds every
	// player's green attackers, whoever controls the source.
	pushTaxEnchantment(g, them, "Flooded Woodlands", sacrificeTaxOracle)
	green1 := pushUnsickCreature(g, me, "Green One", "{G}")
	green2 := pushUnsickCreature(g, me, "Green Two", "{G}")
	white := pushUnsickCreature(g, me, "White One", "{W}")
	l1 := pushTestLand(g, me, "Forest", "Basic Land — Forest")
	l2 := pushTestLand(g, me, "Forest", "Basic Land — Forest")
	advanceIntoStep(t, g, StepDeclareAttackers)

	// The white creature attacks for nothing.
	if err := g.DeclareAttacker(white, them.ID); err != nil {
		t.Fatalf("a nongreen attacker was taxed: %v", err)
	}
	// One land for one green attacker.
	if err := g.DeclareAttackerWith(green1, them.ID, DeclareAttackersParams{
		Pay: CombatCostPayment{SacrificeIDs: []uuid.UUID{l1}},
	}); err != nil {
		t.Fatalf("one green attacker, one land: %v", err)
	}
	// And the second green attacker owes its own land, not none.
	if err := g.DeclareAttacker(green2, them.ID); !errors.Is(err, ErrCombatCostUnpaid) {
		t.Fatalf("the second green attacker = %v, want ErrCombatCostUnpaid", err)
	}
	if err := g.DeclareAttackerWith(green2, them.ID, DeclareAttackersParams{
		Pay: CombatCostPayment{SacrificeIDs: []uuid.UUID{l2}},
	}); err != nil {
		t.Fatalf("the second green attacker with its land: %v", err)
	}
}

// TestAttackCostTapIsPaidByTappingAndNotByAnAttacker — Hollow Warrior:
// the creature tapped to pay cannot be one declared as an attacker this
// combat, and paying taps it.
func TestAttackCostTapIsPaidByTappingAndNotByAnAttacker(t *testing.T) {
	g := newActiveGame(t)
	withCatalogAttackTaxes(t, func(key string) []AttackTax {
		if key != tapTaxOracle {
			return nil
		}
		return []AttackTax{{
			Label:      "This creature can't attack unless you tap an untapped creature you control not declared as an attacking or blocking creature this combat.",
			OnAttacker: true,
			AppliesTo: func(q AttackTaxQuery) bool {
				return q.Attacker.InstanceID == q.Source.InstanceID
			},
			NonMana: CombatCost{
				Tap: &TapOthersCost{
					Count: 1,
					Filter: &TargetSpec{
						Mode: "permanent", Label: "an untapped creature you control", Min: 1, Max: 1,
						Zones: []ZoneKind{ZoneBattlefield},
						CardOK: func(gg *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
							if !c.IsCreature() || c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil {
								return false
							}
							return !gg.DeclaredAsAttackerThisCombat(&c) && !gg.DeclaredAsBlockerThisCombat(&c)
						},
					},
					Label: "an untapped creature you control",
				},
				Label: "an untapped creature you control not declared as an attacking or blocking creature this combat",
			},
		}}
	})
	me, them := g.Seats[0], g.Seats[1]
	warrior := pushKeywordCreature(t, g, me, 4, 4)
	helper := pushKeywordCreature(t, g, me, 1, 1)
	attacker := pushKeywordCreature(t, g, me, 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{warrior, helper, attacker} {
			findBattlefieldCard(g, id).SummonedThisTurn = false
		}
		findBattlefieldCard(g, warrior).OracleID = tapTaxOracle
	})
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(attacker, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// The creature already attacking was declared this combat, so it
	// cannot be tapped for the Warrior.
	err := g.DeclareAttackerWith(warrior, them.ID, DeclareAttackersParams{
		Pay: CombatCostPayment{TapIDs: []uuid.UUID{attacker}},
	})
	if !errors.Is(err, ErrCombatCostUnpaid) {
		t.Fatalf("tapping an attacking creature = %v, want ErrCombatCostUnpaid", err)
	}
	if err := g.DeclareAttackerWith(warrior, them.ID, DeclareAttackersParams{
		Pay: CombatCostPayment{TapIDs: []uuid.UUID{helper}},
	}); err != nil {
		t.Fatalf("tapping the idle creature: %v", err)
	}
	if c := findCard(g, helper); !c.Tapped {
		t.Error("the named creature was not tapped")
	}
	if c := findCard(g, warrior); c.AttackingTarget != them.ID {
		t.Error("the Warrior is not attacking")
	}
}

// TestAttackCostReturnGoesToItsOwnersHand — Floodtide Serpent: the
// enchantment is returned to its OWNER's hand (CR 109.5), and the
// declaration stands.
func TestAttackCostReturnGoesToItsOwnersHand(t *testing.T) {
	g, serpent := taxedTable(t, map[string][]AttackTax{
		returnTaxOracle: {{
			Label:      "This creature can't attack unless you return an enchantment you control to its owner's hand.",
			OnAttacker: true,
			AppliesTo: func(q AttackTaxQuery) bool {
				return q.Attacker.InstanceID == q.Source.InstanceID
			},
			NonMana: CombatCost{
				Return: &ReturnToHandCost{
					Count: 1,
					Filter: &TargetSpec{
						Mode: "permanent", Label: "an enchantment you control", Min: 1, Max: 1,
						Zones: []ZoneKind{ZoneBattlefield},
						CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
							return c.IsEnchantment()
						},
					},
					Label: "an enchantment you control",
				},
				Label: "an enchantment you control",
			},
		}},
	})
	g.WithWriteLock(func() { findBattlefieldCard(g, serpent).OracleID = returnTaxOracle })
	aura := pushTypedTestCard(g, Card{
		Name: "Test Enchantment", TypeLine: "Enchantment",
		Owner: g.Seats[0].ID, Controller: g.Seats[0].ID,
	})
	before := g.Seats[0].Hand.Size()

	if err := g.DeclareAttackerWith(serpent, g.Seats[1].ID, DeclareAttackersParams{
		Pay: CombatCostPayment{ReturnIDs: []uuid.UUID{aura}},
	}); err != nil {
		t.Fatalf("DeclareAttacker returning the enchantment: %v", err)
	}
	if findBattlefieldCard(g, aura) != nil {
		t.Error("the enchantment is still on the battlefield")
	}
	if n := g.Seats[0].Hand.Size(); n != before+1 {
		t.Errorf("hand = %d, want %d — the enchantment went to its owner's hand", n, before+1)
	}
}

// TestAttackCostIsNotFreeSoNoRequirementOwesIt — CR 508.1d: "If a
// creature can't attack unless a player pays a cost, that player is not
// required to pay that cost." The price reads as not free, which is what
// keeps such a creature out of the requirement search.
func TestAttackCostIsNotFreeSoNoRequirementOwesIt(t *testing.T) {
	g, dragon := taxedTable(t, map[string][]AttackTax{
		sacrificeTaxOracle: {sacrificeToAttackTax(landSpec(1), "a land")},
	})
	g.WithWriteLock(func() { findBattlefieldCard(g, dragon).OracleID = sacrificeTaxOracle })
	pushTestLand(g, g.Seats[0], "Plains", "Basic Land — Plains")
	var price AttackTaxPrice
	g.ReadSnapshot(func() {
		price = g.PriceAttackDeclarationForEffect([]AttackDeclaration{{Attacker: dragon, Target: g.Seats[1].ID}})
	})
	if price.IsFree() {
		t.Fatal("a declaration that owes a land reads as free")
	}
	if got := price.Describe(); got != "sacrificing a land" {
		t.Errorf("Describe() = %q", got)
	}
}

// TestBlockCostTapIsPaidAtTheCompletion — CR 509.1d: the cost of a
// block is determined and paid as the defender's declaration completes,
// so staging charges nothing, finishing without naming the creature is
// refused, and finishing with it taps it.
func TestBlockCostTapIsPaidAtTheCompletion(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	prev := CatalogBlockTaxes
	t.Cleanup(func() { CatalogBlockTaxes = prev })
	CatalogBlockTaxes = func(key string) []BlockTax {
		if key != tapTaxOracle {
			return nil
		}
		return []BlockTax{{
			Label:     "This creature can't block unless you tap an untapped creature you control.",
			AppliesTo: BlockedBySelf,
			NonMana: CombatCost{
				Tap: &TapOthersCost{
					Count: 1,
					Filter: &TargetSpec{
						Mode: "permanent", Label: "an untapped creature you control", Min: 1, Max: 1,
						Zones: []ZoneKind{ZoneBattlefield},
						CardOK: func(gg *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
							if !c.IsCreature() || c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil {
								return false
							}
							return !gg.DeclaredAsAttackerThisCombat(&c) && !gg.DeclaredAsBlockerThisCombat(&c)
						},
					},
					Label: "an untapped creature you control",
				},
				Label: "an untapped creature you control",
			},
		}}
	}
	attacker := pushKeywordCreature(t, g, me, 2, 2)
	warrior := pushKeywordCreature(t, g, them, 4, 4)
	helper := pushKeywordCreature(t, g, them, 1, 1)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{attacker, warrior, helper} {
			findBattlefieldCard(g, id).SummonedThisTurn = false
		}
		findBattlefieldCard(g, warrior).OracleID = tapTaxOracle
	})
	attackAndStepToBlocks(t, g, them.ID, attacker)

	// Staging the block charges nothing.
	if err := g.DeclareBlocker(warrior, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if c := findCard(g, helper); c.Tapped {
		t.Error("staging a block paid the cost")
	}
	if err := g.FinishBlocksWith(them.ID, DeclareBlockersParams{}); !errors.Is(err, ErrCombatCostUnpaid) {
		t.Fatalf("finishing without naming a creature = %v, want ErrCombatCostUnpaid", err)
	}
	if c := findCard(g, warrior); c.BlockingTarget != attacker {
		t.Error("a refused completion took the block back")
	}
	if err := g.FinishBlocksWith(them.ID, DeclareBlockersParams{
		Pay: CombatCostPayment{TapIDs: []uuid.UUID{helper}},
	}); err != nil {
		t.Fatalf("finishing with the creature named: %v", err)
	}
	if c := findCard(g, helper); !c.Tapped {
		t.Error("the named creature was not tapped")
	}
	if c := findCard(g, warrior); c.BlockingTarget != attacker {
		t.Error("the paid block was taken back")
	}
}

// TestCombatCostPayableReadsTheWholeBoard — the payability predicate the
// enumerator and the dock share: "a land" plus "two Islands" needs three
// lands, two of them Islands.
func TestCombatCostPayableReadsTheWholeBoard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lines := []CombatCostLine{
		{Label: "a land", Payer: me.ID, Cost: CombatCost{Sacrifice: landSpec(1), Label: "a land"}},
		{Label: "two Islands", Payer: me.ID, Cost: CombatCost{Sacrifice: islandSpec(2), Label: "two Islands"}},
	}
	check := func(want bool, why string) {
		t.Helper()
		var got bool
		g.ReadSnapshot(func() { got = g.CombatCostPayableForEffect(me.ID, lines) })
		if got != want {
			t.Errorf("%s: payable = %v, want %v", why, got, want)
		}
	}
	check(false, "an empty board")
	pushTestLand(g, me, "Island", "Basic Land — Island")
	pushTestLand(g, me, "Forest", "Basic Land — Forest")
	pushTestLand(g, me, "Forest", "Basic Land — Forest")
	check(false, "one Island and two Forests")
	pushTestLand(g, me, "Island", "Basic Land — Island")
	check(true, "two Islands and two Forests")
}
