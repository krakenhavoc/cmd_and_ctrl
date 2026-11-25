package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// partial_cost_change_test.go — #6206: a cost change aimed at PART of a
// cost. Two halves, one issue:
//
//   - "Buyback costs cost {2} less" (Memory Crystal), a reduction of one
//     component of the cost (CR 702.27a, 118.8, 118.7a, 601.2f) — the
//     spell's own mana is untouched, an unbought cast is not discounted
//     at all, and the floor is the component's own generic mana;
//   - "This spell costs 3 life more to cast for each target" (Phyrexian
//     Purge), a per-target life cost the spell charges itself (CR
//     601.2c, 601.2f, 119.4, 601.2h).
//
// What a bug here would hide, in the order it would hurt: a component
// reduction spending itself on the spell's own cost (a Capsize cast for
// less than it prints); a discount on a cast that paid no buyback; a
// life cost the engine charges but the enumerator and the picker do not
// know about; and #289's failure mode, a modifier the engine cannot
// price shipping a spell cheaper than printed instead of refusing it.

const (
	testPartialBuybackOracle = "test-6206-buyback-spell"
	testPartialCrystalOracle = "test-6206-crystal"
	testPartialPurgeOracle   = "test-6206-purge"
)

// buybackSpellInHand seeds a spell with the printed cost `manaCost` and
// one optional buyback cost of `buyback` in the seat's hand, and returns
// its instance ID. A second optional cost may follow it (the kicker in
// the partition test below).
func buybackSpellInHand(t *testing.T, g *Game, me *Player, manaCost, buyback string, extra ...AdditionalCost) uuid.UUID {
	t.Helper()
	costs := []AdditionalCost{{Optional: true, Key: BuybackKey, ManaCost: buyback, Label: "Buyback " + buyback}}
	costs = append(costs, extra...)
	stubOptionalCosts(t, testPartialBuybackOracle, costs)
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Capsize", me.ID)
	c.TypeLine = "Instant"
	c.ManaCost = manaCost
	c.OracleID = testPartialBuybackOracle
	me.Hand.PushTop(c)
	return c.InstanceID
}

// crystalOnBattlefield drops a permanent carrying "Buyback costs cost
// {n} less" under `owner`'s control — the component modifier, built the
// way effects.BuybackCostsLess builds it.
func crystalOnBattlefield(t *testing.T, g *Game, owner *Player, n int) {
	t.Helper()
	withCatalogCostModifiers(t, modifiersFor(testPartialCrystalOracle, CostModifier{
		Kind:               CostReduction,
		Label:              "Buyback costs cost {2} less.",
		Amount:             fixed(n),
		PricesOptionalCost: BuybackKey,
	}))
	modifierSource(t, g, owner, "Test Memory Crystal", testPartialCrystalOracle)
}

// priceErrOf is priceOf for the cases that must REFUSE: the error
// effectiveCostLocked returns, or nil.
func priceErrOf(t *testing.T, g *Game, p *Player, cardID uuid.UUID, params CastSpellParams) error {
	t.Helper()
	card, ok := g.LookupCardForEffect(cardID)
	if !ok {
		t.Fatalf("card %s not found", cardID)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	_, err := g.effectiveCostLocked(p, card, params)
	return err
}

// The acceptance case on the issue: Memory Crystal makes a buyback {3}
// cost {1}, and does not reduce the spell's own cost (CR 118.8d — an
// additional cost is not the mana cost — and 118.7a).
func TestBuybackCostsLessPricesOnlyTheBuybackComponent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := buybackSpellInHand(t, g, me, "{2}{U}", "{3}")

	// Printed: {2}{U} plus the buyback {3} is {5}{U}.
	bought := CastSpellParams{OptionalCosts: []int{0}}
	if got := priceOf(t, g, me, spell, bought); got.Generic != 5 || len(got.Required) != 1 {
		t.Fatalf("unmodified bought-back cost = %+v, want {5}{U}", got)
	}

	crystalOnBattlefield(t, g, me, 2)

	// The buyback {3} is now {1}: {2}{U} plus {1} is {3}{U}.
	if got := priceOf(t, g, me, spell, bought); got.Generic != 3 || len(got.Required) != 1 {
		t.Errorf("bought-back cost under the Crystal = %+v, want {3}{U}", got)
	}
	// And the cast that declines the buyback pays exactly what the card
	// prints: the component the Crystal names is not part of this cast
	// (CR 601.2b).
	if got := priceOf(t, g, me, spell, CastSpellParams{}); got.Generic != 2 || len(got.Required) != 1 {
		t.Errorf("unbought cost under the Crystal = %+v, want {2}{U} — the Crystal discounted a cast that paid no buyback", got)
	}
}

// CR 601.2f and 118.7a: a reduction of part of a cost stops at that
// part's generic mana. {2} less off a buyback {1} leaves {0} and takes
// nothing off the spell — where a plain "spells cost {2} less" would
// have eaten the spell's own generic too.
func TestBuybackCostsLessFloorsAtTheBuybackCost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := buybackSpellInHand(t, g, me, "{2}{U}", "{1}")
	bought := CastSpellParams{OptionalCosts: []int{0}}

	if got := priceOf(t, g, me, spell, bought); got.Generic != 3 {
		t.Fatalf("unmodified bought-back generic = %d, want 3 ({2} + buyback {1})", got.Generic)
	}

	crystalOnBattlefield(t, g, me, 2)
	if got := priceOf(t, g, me, spell, bought); got.Generic != 2 || len(got.Required) != 1 {
		t.Errorf("bought-back cost under the Crystal = %+v, want {2}{U}: the overshoot must not reach the spell's own mana", got)
	}

	// The contrast that makes the point: the SAME {2} as a whole-cost
	// reduction takes the total from {3}{U} to {1}{U}, because by then
	// the component's mana is the spell's (CR 601.2f totals them).
	withCatalogCostModifiers(t, modifiersFor(testPartialCrystalOracle, CostModifier{
		Kind: CostReduction, Label: "Spells cost {2} less to cast.", Amount: fixed(2),
	}))
	if got := priceOf(t, g, me, spell, bought); got.Generic != 1 {
		t.Errorf("whole-cost reduction generic = %d, want 1 — the two passes must not be the same pass", got.Generic)
	}
}

// The partition in the other direction: a modifier aimed at the buyback
// component never prices another component. A kicker is an additional
// cost too, and "buyback costs" is not "kicker costs" (CR 702.33a,
// 702.27a).
func TestBuybackCostsLessDoesNotPriceAKickerCost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := buybackSpellInHand(t, g, me, "{U}", "{3}",
		AdditionalCost{Optional: true, Key: KickerKey, ManaCost: "{2}", Label: "Kicker {2}"})
	crystalOnBattlefield(t, g, me, 2)

	// Kicked only: {U} plus the kicker {2}, untouched.
	if got := priceOf(t, g, me, spell, CastSpellParams{OptionalCosts: []int{1}}); got.Generic != 2 {
		t.Errorf("kicked generic = %d, want 2 — a buyback reduction priced the kicker", got.Generic)
	}
	// Both: the buyback {3} comes down to {1}, the kicker {2} stands.
	if got := priceOf(t, g, me, spell, CastSpellParams{OptionalCosts: []int{0, 1}}); got.Generic != 3 {
		t.Errorf("bought-back and kicked generic = %d, want 3 (buyback {1} + kicker {2})", got.Generic)
	}
}

// A buyback cost that is not mana — "Buyback—Sacrifice a land" (Constant
// Mists), "Buyback—Pay 4 life" (Slaughter) — has no mana component for a
// generic reduction to take (CR 118.7a), and the cast is priced exactly
// as it was.
func TestBuybackCostsLessLeavesANonManaBuybackAlone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	stubOptionalCosts(t, testPartialBuybackOracle, []AdditionalCost{
		{Optional: true, Key: BuybackKey, PayLife: 4, Label: "Buyback—Pay 4 life"},
	})
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Slaughter", me.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{2}{B}"
	c.OracleID = testPartialBuybackOracle
	me.Hand.PushTop(c)
	crystalOnBattlefield(t, g, me, 2)

	got := priceOf(t, g, me, c.InstanceID, CastSpellParams{OptionalCosts: []int{0}})
	if got.Generic != 2 || len(got.Required) != 1 {
		t.Errorf("cost = %+v, want {2}{B}: a life buyback has no mana to reduce", got)
	}
}

// #289's rule, inside the component pass: a modifier that returns a
// nonsensical amount REFUSES the cast rather than shipping the spell at
// a price nobody can justify.
func TestABrokenComponentModifierRefusesTheCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := buybackSpellInHand(t, g, me, "{2}{U}", "{3}")
	withCatalogCostModifiers(t, modifiersFor(testPartialCrystalOracle, CostModifier{
		Kind:               CostReduction,
		Label:              "Buyback costs cost {-1} less.",
		Amount:             fixed(-1),
		PricesOptionalCost: BuybackKey,
	}))
	modifierSource(t, g, me, "Test Broken Crystal", testPartialCrystalOracle)

	if err := priceErrOf(t, g, me, spell, CastSpellParams{OptionalCosts: []int{0}}); !errors.Is(err, ErrCostModifier) {
		t.Errorf("a negative component amount = %v, want ErrCostModifier", err)
	}
	// The unbought cast, which the modifier does not price, is unaffected.
	if err := priceErrOf(t, g, me, spell, CastSpellParams{}); err != nil {
		t.Errorf("the unbought cast is still priceable: %v", err)
	}
}

// A component carrying {X} has no honest price: the reduction a printed
// generic cannot absorb is kept against the cost it was measured against
// (CR 107.3a, 118.7a), and the component's X is not the spell's. No
// printed buyback cost carries {X}; one that did would be refused rather
// than guessed at.
func TestAComponentModifierRefusesACostCarryingX(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := buybackSpellInHand(t, g, me, "{2}{U}", "{X}")
	crystalOnBattlefield(t, g, me, 2)

	if err := priceErrOf(t, g, me, spell, CastSpellParams{OptionalCosts: []int{0}, XValue: 2}); !errors.Is(err, ErrCostModifier) {
		t.Errorf("a buyback {X} under a component modifier = %v, want ErrCostModifier", err)
	}
}

// purgeInHand seeds "This spell costs 3 life more to cast for each
// target. Destroy any number of target creatures." — the card's own
// per-target life surcharge plus a one-or-more creature clause — and
// returns it with the creatures it can name.
func purgeInHand(t *testing.T, g *Game, me *Player, life int) (spell uuid.UUID, creatures []uuid.UUID) {
	t.Helper()
	stubRowDefs(t, map[string]*CardDef{
		testPartialPurgeOracle: {
			LifePerTarget: LifePerTargetCost{
				Life:  life,
				Label: "This spell costs 3 life more to cast for each target.",
			},
		},
	})
	withCatalogTargetSpec(t, func(o string) *TargetSpec {
		if o == testPartialPurgeOracle {
			return &TargetSpec{
				Mode:  "creature",
				Label: "any number of target creatures",
				Zones: []ZoneKind{ZoneBattlefield},
				Min:   1, Max: 0,
				CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsCreature() },
			}
		}
		return nil
	})
	advanceTo(t, g, StepPrecombatMain)
	for range 3 {
		c := NewCard("Test Bear", g.Seats[1].ID)
		c.TypeLine = "Creature — Bear"
		c.Controller = g.Seats[1].ID
		c.Power, c.Toughness = 2, 2
		g.Battlefield.PushTop(c)
		creatures = append(creatures, c.InstanceID)
	}
	s := NewCard("Test Phyrexian Purge", me.ID)
	s.TypeLine = "Sorcery"
	s.ManaCost = "{3}{R}{R}"
	s.OracleID = testPartialPurgeOracle
	me.Hand.PushTop(s)
	return s.InstanceID, creatures
}

func cardRefs(ids ...uuid.UUID) []TargetRef {
	out := make([]TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, TargetRef{Kind: TargetCard, ID: id})
	}
	return out
}

// The acceptance case on the issue: Phyrexian Purge with three targets
// costs 9 life (CR 601.2c then 601.2f, paid at 601.2h).
func TestPerTargetLifeChargesOncePerTarget(t *testing.T) {
	for _, tc := range []struct {
		name    string
		targets int
		want    int
	}{
		{"one target", 1, 3},
		{"two targets", 2, 6},
		{"three targets", 3, 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			spell, creatures := purgeInHand(t, g, me, 3)
			me.Life = 20

			if err := g.CastSpell(me.ID, spell, CastSpellParams{Targets: cardRefs(creatures[:tc.targets]...)}); err != nil {
				t.Fatalf("CastSpell at %d targets: %v", tc.targets, err)
			}
			if got := 20 - me.Life; got != tc.want {
				t.Errorf("life paid = %d, want %d", got, tc.want)
			}
		})
	}
}

// CR 119.4: the caster may pay life only down to 0, so a seat that
// cannot pay the whole surcharge is refused the cast — and a refused
// cast pays nothing and moves nothing.
func TestPerTargetLifeIsRefusedWhenItCannotBePaid(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell, creatures := purgeInHand(t, g, me, 3)
	me.Life = 8

	err := g.CastSpell(me.ID, spell, CastSpellParams{Targets: cardRefs(creatures...)})
	if !errors.Is(err, ErrCantPayTargetTax) {
		t.Fatalf("8 life cannot pay 9: got %v, want ErrCantPayTargetTax", err)
	}
	if me.Life != 8 || !me.Hand.Contains(spell) {
		t.Errorf("a refused cast pays and moves nothing: life %d, in hand %v", me.Life, me.Hand.Contains(spell))
	}
	// Two targets cost 6, which 8 life can pay.
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Targets: cardRefs(creatures[:2]...)}); err != nil {
		t.Fatalf("8 life pays 6: %v", err)
	}
	if me.Life != 2 {
		t.Errorf("life = %d, want 2", me.Life)
	}
}

// CR 119.4 again: exactly the life total pays, as it does for every
// other life cost of a cast.
func TestPerTargetLifeMayBePaidExactly(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell, creatures := purgeInHand(t, g, me, 3)
	me.Life = 9

	if err := g.CastSpell(me.ID, spell, CastSpellParams{Targets: cardRefs(creatures...)}); err != nil {
		t.Fatalf("9 life pays 9: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life = %d, want 0", me.Life)
	}
}

// The surcharge is summed with the cast's OTHER life costs before CR
// 119.4 is asked, the way the commander tax in life and the Phyrexian
// symbols are: 6 for two targets plus a buyback's "pay 4 life" is 10,
// and a seat at 9 is refused although either cost alone would pass.
func TestPerTargetLifeIsSummedWithTheOtherLifeCostsOfTheCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell, creatures := purgeInHand(t, g, me, 3)
	stubOptionalCosts(t, testPartialPurgeOracle, []AdditionalCost{
		{Optional: true, Key: BuybackKey, PayLife: 4, Label: "Buyback—Pay 4 life"},
	})
	me.Life = 9

	two := CastSpellParams{Targets: cardRefs(creatures[:2]...), OptionalCosts: []int{0}}
	if err := g.CastSpell(me.ID, spell, two); !errors.Is(err, ErrCantPayTargetTax) {
		t.Fatalf("9 life cannot pay 6 + 4: got %v, want ErrCantPayTargetTax", err)
	}
	if me.Life != 9 {
		t.Errorf("a refused cast pays nothing: life = %d", me.Life)
	}
	me.Life = 10
	if err := g.CastSpell(me.ID, spell, two); err != nil {
		t.Fatalf("10 life pays 6 + 4: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life = %d, want 0 (6 for the targets, 4 for the buyback)", me.Life)
	}
}

// "For each target" counts TARGETS (CR 601.2c), not targeted
// permanents: a clause that names a player pays for it too, which is
// the number Fireball's "for each target beyond the first" counts.
func TestPerTargetLifeCountsEveryTargetNotJustPermanents(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell, creatures := purgeInHand(t, g, me, 3)
	withCatalogTargetSpec(t, func(o string) *TargetSpec {
		if o == testPartialPurgeOracle {
			return &TargetSpec{
				Mode:    "any",
				Label:   "any number of targets",
				Zones:   []ZoneKind{ZoneBattlefield},
				Players: true,
				Min:     1, Max: 0,
				CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsCreature() },
			}
		}
		return nil
	})
	me.Life = 20

	targets := append(cardRefs(creatures[0]), TargetRef{Kind: TargetPlayer, ID: opp.ID})
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell at a creature and a player: %v", err)
	}
	if got := 20 - me.Life; got != 6 {
		t.Errorf("life paid = %d, want 6 (a creature and a player are two targets)", got)
	}
}

// The breakdown a refusal quotes: the spell's own line names no
// permanent, so the sentence does not claim one is being targeted.
func TestPerTargetLifeRefusalReadsAsTheSpellsOwnCost(t *testing.T) {
	lines := []TargetTaxLine{{
		SourceName: "Phyrexian Purge",
		Label:      "This spell costs 3 life more to cast for each target.",
		Life:       9,
		Self:       true,
	}}
	err := targetTaxRefusal(lines, 9)
	if !errors.Is(err, ErrCantPayTargetTax) {
		t.Fatalf("refusal = %v, want ErrCantPayTargetTax", err)
	}
	if got := err.Error(); !strings.Contains(got, "costs 9 more life") || strings.Contains(got, "targeting Phyrexian Purge") {
		t.Errorf("refusal reads %q, want the spell's own sentence", got)
	}
}
