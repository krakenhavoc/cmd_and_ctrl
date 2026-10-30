package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// target_total_test.go — #4616, the engine half of the total set rule:
// "up to two target creatures with total mana value 6 or less" (The
// Super Hero Civil War, chapter I). Its readers are the announce gate
// (CR 601.2c), the CR 608.2b re-check over the survivors, the CR 603.3d
// / cast-offer fillability count, and the CR 115.7 retarget offer and
// gate.

const (
	totalOracle      = "test-total-up-to-two-creatures-total-mana-value-6"
	totalExactOracle = "test-total-exactly-two-creatures-total-mana-value-6"
)

// totalManaValueSpec is "<min> to <max> target creatures with total
// mana value 6 or less", written against the engine alone.
func totalManaValueSpec(min, max int) *TargetSpec {
	return (&TargetSpec{
		Mode:  "creature",
		Label: "up to two target creatures with total mana value 6 or less",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: min, Max: max,
	}).WithTotal(&TargetTotal{Label: "have total mana value 6 or less", Stat: TargetTotalManaValue, Max: 6})
}

func totalSpecs(t *testing.T) {
	t.Helper()
	withCatalogTargetSpec(t, func(oracleID string) *TargetSpec {
		switch oracleID {
		case totalOracle:
			return totalManaValueSpec(0, 2)
		case totalExactOracle:
			return totalManaValueSpec(2, 2)
		}
		return nil
	})
}

// pushTotalCreature puts a creature with mana cost `cost` onto the
// battlefield under owner.
func pushTotalCreature(g *Game, owner *Player, name, cost string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Test"
	c.ManaCost = cost
	c.Power, c.Toughness = 2, 2
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func totalCast(t *testing.T, g *Game, caster *Player, oracle string, targets ...uuid.UUID) (*StackItem, error) {
	t.Helper()
	c := NewCard("Total Spell", caster.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{0}"
	c.OracleID = oracle
	c.Controller = caster.ID
	caster.Hand.PushTop(c)
	refs := make([]TargetRef, 0, len(targets))
	for _, id := range targets {
		refs = append(refs, TargetRef{Kind: TargetCard, ID: id})
	}
	if err := g.CastSpell(caster.ID, c.InstanceID, CastSpellParams{Targets: refs}); err != nil {
		return nil, err
	}
	return g.StackMeta[c.InstanceID], nil
}

// setBattlefieldCost rewrites a permanent's mana cost in place — the
// test's stand-in for a copy effect or a face turning, which is how a
// creature's mana value changes on the battlefield.
func setBattlefieldCost(g *Game, id uuid.UUID, cost string) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].ManaCost = cost
		}
	}
}

// CR 601.2c: a set over the total is refused with the printed rule in
// the refusal; a set at the total, one card, and nothing are fine.
func TestTotalRefusesASetOverTheTotalAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	totalSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	four := pushTotalCreature(g, opp, "Four", "{3}{G}")
	three := pushTotalCreature(g, opp, "Three", "{2}{G}")
	two := pushTotalCreature(g, opp, "Two", "{1}{G}")
	seven := pushTotalCreature(g, opp, "Seven", "{6}{G}")

	_, err := totalCast(t, g, me, totalOracle, four, three)
	if !errors.Is(err, ErrIllegalTarget) {
		t.Fatalf("mana values 4 and 3: err = %v, want ErrIllegalTarget", err)
	}
	if !strings.Contains(err.Error(), "have total mana value 6 or less") {
		t.Errorf("the refusal names the rule: %q", err.Error())
	}
	if _, err := totalCast(t, g, me, totalOracle, seven); !errors.Is(err, ErrIllegalTarget) {
		t.Errorf("one creature over the total on its own: err = %v, want ErrIllegalTarget", err)
	}
	if _, err := totalCast(t, g, me, totalOracle, four, two); err != nil {
		t.Fatalf("mana values 4 and 2 total exactly 6 and are legal: %v", err)
	}
	if _, err := totalCast(t, g, me, totalOracle, three); err != nil {
		t.Fatalf("one creature within the total is legal: %v", err)
	}
	if _, err := totalCast(t, g, me, totalOracle); err != nil {
		t.Fatalf("\"up to\" with nothing chosen is legal: %v", err)
	}
}

// A card whose cost the engine cannot read has no mana value to add, so
// it fits no total — as it meets no "mana value N or less".
func TestTotalRefusesACardWithNoReadableManaValue(t *testing.T) {
	g := newActiveGame(t)
	totalSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	odd := pushTotalCreature(g, opp, "Odd", "{1}{R} // {1}{U}")
	if _, err := totalCast(t, g, me, totalOracle, odd); !errors.Is(err, ErrIllegalTarget) {
		t.Errorf("an unreadable cost: err = %v, want ErrIllegalTarget", err)
	}
}

// CR 608.2b: the survivors are summed again on resolution. Within the
// total both are legal; when one's mana value grows so the sum passes
// 6, BOTH are illegal and the spell has no legal target left; and when
// one then leaves, the other is summed alone and is legal again.
func TestTotalIsRejudgedOverTheSurvivorsAtResolution(t *testing.T) {
	g := newActiveGame(t)
	totalSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushTotalCreature(g, opp, "A", "{2}{G}")
	b := pushTotalCreature(g, opp, "B", "{2}{G}")
	item, err := totalCast(t, g, me, totalOracle, a, b)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	ra, rb := item.Targets[0], item.Targets[1]
	g.WithWriteLock(func() {
		if !g.TargetStillLegalForEffect(item, ra) || !g.TargetStillLegalForEffect(item, rb) {
			t.Error("two creatures totalling 6 are both legal")
		}
		setBattlefieldCost(g, b, "{4}{G}")
		if g.TargetStillLegalForEffect(item, ra) || g.TargetStillLegalForEffect(item, rb) {
			t.Error("survivors totalling 8 are all illegal — neither is preferred")
		}
		if !spellAllTargetsIllegalLocked(g, item) {
			t.Error("with both illegal the spell has no legal target left")
		}
		g.Battlefield.Remove(b)
		if !g.TargetStillLegalForEffect(item, ra) {
			t.Error("a target that left adds nothing; the survivor is within the total again")
		}
	})
}

// A survivor whose cost can no longer be read fits no total and is
// illegal itself; the other is summed without it.
func TestTotalSurvivorWithNoReadableValueIsIllegalAlone(t *testing.T) {
	g := newActiveGame(t)
	totalSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushTotalCreature(g, opp, "A", "{2}{G}")
	b := pushTotalCreature(g, opp, "B", "{1}{G}")
	item, err := totalCast(t, g, me, totalOracle, a, b)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		setBattlefieldCost(g, b, "{1}{R} // {1}{U}")
		if g.TargetStillLegalForEffect(item, item.Targets[1]) {
			t.Error("a survivor with no readable mana value fits no total")
		}
		if !g.TargetStillLegalForEffect(item, item.Targets[0]) {
			t.Error("the other survivor is summed alone and stays legal")
		}
	})
}

// An exact count is fillable only when that many fit within the total
// together (CR 603.3d, and the cast offer): the smallest values first.
func TestTotalCountsTheSmallestValuesForFillability(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	pushTotalCreature(g, opp, "Five", "{4}{G}")
	pushTotalCreature(g, opp, "Four", "{3}{G}")
	pushTotalCreature(g, opp, "Odd", "{1}{R} // {1}{U}")
	steps := AnnouncedClauses(totalManaValueSpec(2, 2), nil, nil)
	g.WithWriteLock(func() {
		if !g.anyClauseUnfillableLocked(SourceChooser(me.ID), steps) {
			t.Error("mana values 5 and 4 (and an unreadable cost) are no legal pair")
		}
	})
	pushTotalCreature(g, opp, "Two", "{1}{G}")
	g.WithWriteLock(func() {
		if g.anyClauseUnfillableLocked(SourceChooser(me.ID), steps) {
			t.Error("a 2 makes the 4 and the 2 a legal pair")
		}
	})
}

// CR 115.7d / 115.7e: the one-slot offer narrows to the cards that fit
// beside the staying slot, and a slot moved alone past the total is
// refused — but the gate judges only the final set, so moving both
// slots at once to a pair within the total is legal.
func TestRetargetRespectsTheTotalRule(t *testing.T) {
	g := newActiveGame(t)
	totalSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushTotalCreature(g, opp, "A", "{2}{G}")
	b := pushTotalCreature(g, opp, "B", "{2}{G}")
	one := pushTotalCreature(g, opp, "One", "{G}")
	five := pushTotalCreature(g, opp, "Five", "{4}{G}")
	item, err := totalCast(t, g, me, totalOracle, a, b)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: me.ID, Policy: RetargetChooseNew, Optional: true,
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	first := findRetargetPrompt(g, me.ID)
	if first == nil || first.RetargetSlot != 0 {
		t.Fatalf("first prompt = %+v, want slot 0", first)
	}
	if !hasUUID(first.PickTargetCards, one) {
		t.Error("a 1 fits beside the staying 3 and is a legal new target")
	}
	if hasUUID(first.PickTargetCards, five) {
		t.Error("a 5 beside the staying 3 is over the total and is not offered")
	}
	if err := g.ResolveRetarget(first.ID, me.ID, []TargetRef{{Kind: TargetCard, ID: five}}); !errors.Is(err, ErrIllegalTarget) {
		t.Errorf("moving one slot past the total is refused, err = %v", err)
	}

	// The whole list at once: the 5 and the 1, which total 6.
	next := append([]TargetRef(nil), item.Targets...)
	next[0].ID, next[1].ID = five, one
	g.WithWriteLock(func() {
		if err := g.RetargetStackItemForEffect(item.ID, me.ID, RetargetChooseNew, next); err != nil {
			t.Fatalf("moving both slots to a pair within the total is legal: %v", err)
		}
	})
	if item.Targets[0].ID != five || item.Targets[1].ID != one {
		t.Errorf("targets = %v, want Five and One", item.Targets)
	}
}

// The values the view and the enumerator read: each card's mana value,
// none for an unreadable cost, and no map for a spec without the rule.
func TestTotalValuesAreTheManaValues(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1]
	three := pushTotalCreature(g, opp, "Three", "{2}{G}")
	token := pushTotalCreature(g, opp, "Token", "")
	odd := pushTotalCreature(g, opp, "Odd", "{1}{R} // {1}{U}")
	var vals map[uuid.UUID]int
	var none map[uuid.UUID]int
	g.WithWriteLock(func() {
		vals = g.TargetTotalValuesForEffect(totalManaValueSpec(0, 2), []uuid.UUID{three, token, odd})
		none = g.TargetTotalValuesForEffect(&TargetSpec{Zones: []ZoneKind{ZoneBattlefield}}, []uuid.UUID{three})
	})
	if vals[three] != 3 {
		t.Errorf("value of a {2}{G} = %d, want 3", vals[three])
	}
	if v, ok := vals[token]; !ok || v != 0 {
		t.Errorf("a card with no mana cost has mana value 0 (CR 202.3a), got %d, %v", v, ok)
	}
	if _, ok := vals[odd]; ok {
		t.Error("an unreadable cost has no value")
	}
	if none != nil {
		t.Errorf("a spec without the rule has no values, got %v", none)
	}
	// CR 708.2a: a face-down permanent has no mana cost, so it adds 0,
	// whatever the card it hides costs.
	hidden := pushTotalCreature(g, opp, "Hidden", "{6}{G}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == hidden {
				g.Battlefield.Cards[i].FaceDown = true
				g.Battlefield.Cards[i].FaceDownKind = FaceDownManifested
			}
		}
		vals = g.TargetTotalValuesForEffect(totalManaValueSpec(0, 2), []uuid.UUID{hidden})
	})
	if v, ok := vals[hidden]; !ok || v != 0 {
		t.Errorf("a face-down permanent's value = %d, %v, want 0 (it has no mana cost)", v, ok)
	}
	if _, ok := (&TargetTotal{Label: "x"}).valueOf(g, Card{ManaCost: "{2}"}); ok {
		t.Error("a rule with no statistic sums nothing")
	}
}
