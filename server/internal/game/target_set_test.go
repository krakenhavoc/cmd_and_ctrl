package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// target_set_test.go — #1559, the engine half of a rule over the
// chosen SET of targets (CR 601.2c): the announce gate, the CR 608.2b
// re-check over the survivors, the retarget offer and gate (CR 115.7c),
// the CR 603.3d / cast-offer fillability count, and the X bound.

const setRuleOracle = "test-set-rule-two-creatures-different-controllers"

// differentControllersSpec is Run Away Together's clause written
// against the engine alone: two target creatures, no two sharing a
// controller.
func differentControllersSpec() *TargetSpec {
	return (&TargetSpec{
		Mode:  "creature",
		Label: "two target creatures controlled by different players",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 2, Max: 2,
	}).EachDifferent(&TargetDifference{
		Label: "be controlled by different players",
		Key: func(_ *Game, c Card, _ ZoneKind) (string, bool) {
			return c.Controller.String(), true
		},
	})
}

func setRuleSpecs(t *testing.T) {
	t.Helper()
	withCatalogTargetSpec(t, func(oracleID string) *TargetSpec {
		if oracleID == setRuleOracle {
			return differentControllersSpec()
		}
		return nil
	})
}

func setRuleCast(t *testing.T, g *Game, caster *Player, targets ...uuid.UUID) (*StackItem, error) {
	t.Helper()
	c := NewCard("Pair Spell", caster.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{0}"
	c.OracleID = setRuleOracle
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

func TestSetRuleRefusesASetThatBreaksItAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	setRuleSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	b := pushRetargetCreature(g, opp, "B")
	mine := pushRetargetCreature(g, me, "Mine")

	_, err := setRuleCast(t, g, me, a, b)
	if !errors.Is(err, ErrIllegalTarget) {
		t.Fatalf("two creatures of one controller: err = %v, want ErrIllegalTarget", err)
	}
	if !strings.Contains(err.Error(), "be controlled by different players") {
		t.Errorf("the refusal names the rule: %q", err.Error())
	}
	if _, err := setRuleCast(t, g, me, a, mine); err != nil {
		t.Fatalf("a pair with different controllers is legal: %v", err)
	}
}

// CR 608.2b over the survivors: one target leaving costs only that
// target; two survivors that now share a key are BOTH illegal.
func TestSetRuleIsRejudgedOverTheSurvivorsAtResolution(t *testing.T) {
	g := newActiveGame(t)
	setRuleSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushRetargetCreature(g, opp, "A")
	mine := pushRetargetCreature(g, me, "Mine")
	item, err := setRuleCast(t, g, me, a, mine)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	ra, rm := item.Targets[0], item.Targets[1]

	// Both still held by different players: both legal.
	g.WithWriteLock(func() {
		if !g.TargetStillLegalForEffect(item, ra) || !g.TargetStillLegalForEffect(item, rm) {
			t.Error("an intact pair is legal")
		}
	})
	// Mine moves to the opponent: the survivors share a controller.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mine {
				g.Battlefield.Cards[i].Controller = opp.ID
				g.Battlefield.Cards[i].BaseController = opp.ID
			}
		}
		if g.TargetStillLegalForEffect(item, ra) || g.TargetStillLegalForEffect(item, rm) {
			t.Error("two survivors that share a key are both illegal — neither is preferred")
		}
		if !spellAllTargetsIllegalLocked(g, item) {
			t.Error("with both illegal the spell has no legal target left")
		}
	})
	// One of them leaves: the other no longer conflicts with anything.
	g.WithWriteLock(func() {
		g.Battlefield.Remove(mine)
		if !g.TargetStillLegalForEffect(item, ra) {
			t.Error("a target that left conflicts with nothing; the survivor is legal again")
		}
	})
}

// A clause with a set rule and Min 2 whose candidates all share one
// key is unfillable (CR 603.3d, and the cast offer), even though it
// has two candidates.
func TestSetRuleCountsKeysNotCandidatesForFillability(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	pushRetargetCreature(g, opp, "A")
	pushRetargetCreature(g, opp, "B")
	steps := AnnouncedClauses(differentControllersSpec(), nil, nil)
	g.WithWriteLock(func() {
		if !g.anyClauseUnfillableLocked(SourceChooser(me.ID), steps) {
			t.Error("two creatures under one controller are no legal pair")
		}
	})
	pushRetargetCreature(g, me, "Mine")
	g.WithWriteLock(func() {
		if g.anyClauseUnfillableLocked(SourceChooser(me.ID), steps) {
			t.Error("a creature under a second controller makes the pair fillable")
		}
	})
}

// CR 115.7c: a retarget may not move a slot onto a key another slot
// holds. The offer leaves such a card out, and the gate refuses it.
func TestRetargetRespectsTheSetRule(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	setRuleSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	a := pushRetargetCreature(g, opp, "A")
	mine := pushRetargetCreature(g, me, "Mine")
	mine2 := pushRetargetCreature(g, me, "Mine Too")
	theirs := pushRetargetCreature(g, third, "Theirs")
	item, err := setRuleCast(t, g, me, a, mine)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID: item.ID, Chooser: opp.ID, Policy: RetargetChooseNew, Optional: true,
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	first := findRetargetPrompt(g, opp.ID)
	if first == nil || first.RetargetSlot != 0 {
		t.Fatalf("first prompt = %+v, want slot 0", first)
	}
	if hasUUID(first.PickTargetCards, mine2) {
		t.Error("slot 0 may not move onto a creature of the controller slot 1 already holds")
	}
	if !hasUUID(first.PickTargetCards, theirs) {
		t.Error("a creature of a third controller is a legal new target")
	}
	if err := g.ResolveRetarget(first.ID, opp.ID, []TargetRef{{Kind: TargetCard, ID: mine2}}); !errors.Is(err, ErrIllegalTarget) {
		t.Errorf("a retarget that breaks the set rule is refused, err = %v", err)
	}
}

// ManaValueAtMostX: unbound (a hand snapshot) admits every card; bound
// to X admits mana value X or less; an unreadable cost meets no bound.
func TestManaValueAtMostXBindsToTheAnnouncedX(t *testing.T) {
	spec := (&TargetSpec{Zones: []ZoneKind{ZoneGraveyard}, Min: 0, Max: 0}).WithManaValueAtMostX()
	three := Card{ManaCost: "{2}{B}"}
	one := Card{ManaCost: "{B}"}
	if !spec.xBoundAdmits(three) {
		t.Error("before X is bound, the clause admits every card")
	}
	steps := AnnouncedClauses(spec, nil, nil)
	bindStepsX(steps, 2)
	if steps[0].Clause.xBoundAdmits(three) || !steps[0].Clause.xBoundAdmits(one) {
		t.Error("bound to X=2: mana value 1 qualifies, 3 does not")
	}
	if spec.xBoundSet {
		t.Error("binding writes the announcement's copy, never the catalog spec")
	}
	if steps[0].Clause.xBoundAdmits(Card{ManaCost: "{1}{R} // {1}{U}"}) {
		t.Error("an unreadable cost meets no bound")
	}
}
