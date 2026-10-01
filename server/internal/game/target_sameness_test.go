package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// target_sameness_test.go — #1807 (ADR 0106 §5), the engine half of
// "targets from a single graveyard": a set rule whose picks must all
// SHARE one key, the card's owner. The four readers are the announce
// gate (CR 601.2c), the CR 608.2b re-check over the survivors, the
// CR 603.3d / cast-offer fillability count, and the CR 115.7 retarget
// offer and gate.

const (
	samenessOracle      = "test-sameness-up-to-four-from-a-single-graveyard"
	samenessExactOracle = "test-sameness-exactly-two-from-a-single-graveyard"
)

// singleGraveyardSpec is "<n> target cards from a single graveyard",
// written against the engine alone.
func singleGraveyardSpec(min, max int) *TargetSpec {
	return (&TargetSpec{
		Mode:  "card_in_graveyard",
		Label: "target cards from a single graveyard",
		Zones: []ZoneKind{ZoneGraveyard},
		CardOK: func(_ *Game, _ uuid.UUID, _ Card, _ ZoneKind) bool {
			return true
		},
		Min: min, Max: max,
	}).AllShare(&TargetSameness{Label: "come from a single graveyard", Key: TargetShareOwner})
}

func samenessSpecs(t *testing.T) {
	t.Helper()
	withCatalogTargetSpec(t, func(oracleID string) *TargetSpec {
		switch oracleID {
		case samenessOracle:
			return singleGraveyardSpec(0, 4)
		case samenessExactOracle:
			return singleGraveyardSpec(2, 2)
		}
		return nil
	})
}

func pushSamenessCard(owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Sorcery"
	owner.Graveyard.PushTop(c)
	return c.InstanceID
}

func samenessCast(t *testing.T, g *Game, caster *Player, oracle string, targets ...uuid.UUID) (*StackItem, error) {
	t.Helper()
	c := NewCard("Graveyard Spell", caster.ID)
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

// CR 601.2c: a set that reaches two graveyards is refused, with the
// printed rule in the refusal; any number from one graveyard is fine.
func TestSamenessRefusesTwoGraveyardsAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	samenessSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	theirs1 := pushSamenessCard(opp, "Theirs 1")
	theirs2 := pushSamenessCard(opp, "Theirs 2")
	mine := pushSamenessCard(me, "Mine")

	_, err := samenessCast(t, g, me, samenessOracle, theirs1, mine)
	if !errors.Is(err, ErrIllegalTarget) {
		t.Fatalf("cards from two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	if !strings.Contains(err.Error(), "come from a single graveyard") {
		t.Errorf("the refusal names the rule: %q", err.Error())
	}
	if _, err := samenessCast(t, g, me, samenessOracle, theirs1, theirs2); err != nil {
		t.Fatalf("two cards from one graveyard are legal: %v", err)
	}
	if _, err := samenessCast(t, g, me, samenessOracle); err != nil {
		t.Fatalf("\"up to\" with nothing chosen is legal: %v", err)
	}
}

// CR 608.2b: a card that left the graveyard in response costs only
// itself; the survivors still share a graveyard and stay legal.
func TestSamenessIsRejudgedOverTheSurvivorsAtResolution(t *testing.T) {
	g := newActiveGame(t)
	samenessSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushSamenessCard(opp, "A")
	b := pushSamenessCard(opp, "B")
	item, err := samenessCast(t, g, me, samenessOracle, a, b)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	ra, rb := item.Targets[0], item.Targets[1]
	g.WithWriteLock(func() {
		if !g.TargetStillLegalForEffect(item, ra) || !g.TargetStillLegalForEffect(item, rb) {
			t.Error("two cards still in one graveyard are both legal")
		}
		opp.Graveyard.Remove(b)
		if g.TargetStillLegalForEffect(item, rb) {
			t.Error("a card that left the graveyard is illegal")
		}
		if !g.TargetStillLegalForEffect(item, ra) {
			t.Error("the survivor conflicts with nothing and stays legal")
		}
	})
}

// The re-check branch is the general one: two survivors whose keys
// came to differ are both illegal (the weaker reading). An owner key
// cannot move in play, so the test moves it by hand.
func TestSamenessSurvivorsThatDisagreeAreBothIllegal(t *testing.T) {
	g := newActiveGame(t)
	samenessSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushSamenessCard(opp, "A")
	b := pushSamenessCard(opp, "B")
	item, err := samenessCast(t, g, me, samenessOracle, a, b)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		for i := range opp.Graveyard.Cards {
			if opp.Graveyard.Cards[i].InstanceID == b {
				opp.Graveyard.Cards[i].Owner = me.ID
			}
		}
		if g.TargetStillLegalForEffect(item, item.Targets[0]) || g.TargetStillLegalForEffect(item, item.Targets[1]) {
			t.Error("two survivors with different keys are both illegal — neither is preferred")
		}
	})
}

// An exact count is fillable only when ONE graveyard holds that many
// (CR 603.3d, and the cast offer): one card in each of two graveyards
// is no legal pair.
func TestSamenessCountsTheLargestGroupForFillability(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	pushSamenessCard(opp, "Theirs")
	pushSamenessCard(me, "Mine")
	steps := AnnouncedClauses(singleGraveyardSpec(2, 2), nil, nil)
	g.WithWriteLock(func() {
		if !g.anyClauseUnfillableLocked(SourceChooser(me.ID), steps) {
			t.Error("one card in each of two graveyards is no legal pair")
		}
	})
	pushSamenessCard(opp, "Theirs Too")
	g.WithWriteLock(func() {
		if g.anyClauseUnfillableLocked(SourceChooser(me.ID), steps) {
			t.Error("a second card in one graveyard makes the pair fillable")
		}
	})
}

// CR 115.7d / 115.7e: the one-slot offer narrows to the graveyard the
// staying slots are in, and a slot moved alone to another graveyard is
// refused — but the gate judges only the final set, so moving EVERY
// slot to another graveyard at once is legal.
func TestRetargetRespectsTheSamenessRule(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	samenessSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	a := pushSamenessCard(opp, "A")
	b := pushSamenessCard(opp, "B")
	c := pushSamenessCard(opp, "C")
	x := pushSamenessCard(third, "X")
	y := pushSamenessCard(third, "Y")
	item, err := samenessCast(t, g, me, samenessOracle, a, b)
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
	if !hasUUID(first.PickTargetCards, c) {
		t.Error("another card of the same graveyard is a legal new target")
	}
	if hasUUID(first.PickTargetCards, x) {
		t.Error("slot 0 alone may not move to a graveyard slot 1 is not in")
	}
	if err := g.ResolveRetarget(first.ID, me.ID, []TargetRef{{Kind: TargetCard, ID: x}}); !errors.Is(err, ErrIllegalTarget) {
		t.Errorf("moving one slot to another graveyard is refused, err = %v", err)
	}

	// The whole list at once: both slots to the third player's
	// graveyard. Only the final set is judged (CR 115.7e).
	next := append([]TargetRef(nil), item.Targets...)
	next[0].ID, next[1].ID = x, y
	g.WithWriteLock(func() {
		if err := g.RetargetStackItemForEffect(item.ID, me.ID, RetargetChooseNew, next); err != nil {
			t.Fatalf("moving every slot to one other graveyard is legal: %v", err)
		}
	})
	if item.Targets[0].ID != x || item.Targets[1].ID != y {
		t.Errorf("targets = %v, want X and Y", item.Targets)
	}
}

// The keys the view and the enumerator read: each card's owner, and
// nothing for a spec without the rule.
func TestSamenessKeysAreTheOwner(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushSamenessCard(me, "Mine")
	theirs := pushSamenessCard(opp, "Theirs")
	var keys, none map[uuid.UUID]string
	g.WithWriteLock(func() {
		keys = g.TargetSamenessKeysForEffect(singleGraveyardSpec(0, 4), []uuid.UUID{mine, theirs})
		none = g.TargetSamenessKeysForEffect(&TargetSpec{Zones: []ZoneKind{ZoneGraveyard}}, []uuid.UUID{mine})
	})
	if keys[mine] != me.ID.String() || keys[theirs] != opp.ID.String() {
		t.Errorf("keys = %v, want each card's owner", keys)
	}
	if none != nil {
		t.Errorf("a spec without the rule has no keys, got %v", none)
	}
	if _, ok := (&TargetSameness{Label: "x"}).keyOf(Card{Owner: me.ID}); ok {
		t.Error("a rule with no key constrains nothing")
	}
}
