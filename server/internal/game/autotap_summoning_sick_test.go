package game

import (
	"testing"

	"github.com/google/uuid"
)

// autotap_summoning_sick_test.go pins issue #540: "the delighted
// halfling came in without summoning sickness". The manual path
// (ActivateManaAbility) has enforced CR 302.6 on creature {T}
// abilities since #233, but the auto-tapper had no such check in
// either half — the planner offered a creature that entered this
// turn as a mana source, and the executor tapped it directly rather
// than going through the activation verb. So paying a cost by
// auto-tap bypassed the rule the hand-click path enforces.
//
// The exceptions are the printed ones: haste, and a permanent that
// has been under its controller's control since the turn began.
// Non-creature sources — rocks, Treasures, lands — are never sick.

// halflingAbilities mirrors Delighted Halfling's printed pair: a
// painless colourless half the auto-tapper will plan, and a
// restricted colour half it already skips (autoTapAbilityFor drops
// restricted output). Only the first matters to these tests; the
// second is here so the card under test is the reported card.
func halflingAbilities() []ManaAbilityShape {
	return []ManaAbilityShape{
		{TapCost: true, Produced: "{C}", Label: "Add {C}"},
		{
			TapCost:      true,
			Produced:     "{W|U|B|R|G}",
			Label:        "Add one mana of any color (legendary spells only)",
			Restrictions: []string{"cast", "supertype:Legendary"},
		},
	}
}

// TestAutoTapSkipsSummoningSickManaCreature is the reported bug: a
// Delighted Halfling that entered this turn must not be planned as a
// mana source.
func TestAutoTapSkipsSummoningSickManaCreature(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := pushIntrinsicPermanent(g, p, "Delighted Halfling", "Creature — Halfling Citizen", halflingAbilities(), nil)
	setSummonedThisTurn(g, id, true)

	if plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{1}"), 0); ok {
		t.Errorf("planned %v off a creature that entered this turn", plan)
	}
}

// Haste exempts it, exactly as on the activation path.
func TestAutoTapHasteBeatsSummoningSickness(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := pushIntrinsicPermanent(g, p, "Hasty Halfling", "Creature — Halfling Citizen",
		halflingAbilities(), []string{"haste"})
	setSummonedThisTurn(g, id, true)

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{1}"), 0)
	if !ok {
		t.Fatalf("haste should let the auto-tapper plan it")
	}
	if len(plan) != 1 || plan[0] != id {
		t.Errorf("plan = %v, want [%v]", plan, id)
	}
}

// A creature that has been out since before this turn is a normal
// mana source — do not over-correct.
func TestAutoTapPlansSettledManaCreature(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := pushIntrinsicPermanent(g, p, "Delighted Halfling", "Creature — Halfling Citizen", halflingAbilities(), nil)
	setSummonedThisTurn(g, id, false)

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{1}"), 0)
	if !ok {
		t.Fatalf("a settled mana creature should be plannable")
	}
	if len(plan) != 1 || plan[0] != id {
		t.Errorf("plan = %v, want [%v]", plan, id)
	}
}

// CR 302.6 is about creatures and nothing else: a rock or a land that
// arrived this turn taps fine.
func TestAutoTapNoncreatureSourcesIgnoreSummoningSickness(t *testing.T) {
	for _, tc := range []struct{ name, typeLine string }{
		{"Sol Ring", "Artifact"},
		{"Ancient Den", "Land"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p := g.Seats[0]
			id := pushIntrinsicPermanent(g, p, tc.name, tc.typeLine,
				[]ManaAbilityShape{{TapCost: true, Produced: "{C}", Label: "Add {C}"}}, nil)
			setSummonedThisTurn(g, id, true)

			plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{1}"), 0)
			if !ok {
				t.Fatalf("a noncreature source is never summoning-sick")
			}
			if len(plan) != 1 || plan[0] != id {
				t.Errorf("plan = %v, want [%v]", plan, id)
			}
		})
	}
}

// The planner routes around the sick creature rather than failing
// outright when another source can pay.
func TestAutoTapRoutesAroundSummoningSickCreature(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	sick := pushIntrinsicPermanent(g, p, "Delighted Halfling", "Creature — Halfling Citizen", halflingAbilities(), nil)
	setSummonedThisTurn(g, sick, true)
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")

	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{1}"), 0)
	if !ok {
		t.Fatalf("the Forest can pay {1}")
	}
	if len(plan) != 1 || plan[0] != forest {
		t.Errorf("plan = %v, want [%v] (the Forest, not the sick Halfling)", plan, forest)
	}
}

// The executor is the second half of the fix. It taps `card.Tapped =
// true` directly instead of routing through ActivateManaAbility, so a
// planner-only fix leaves a live path to the same illegal tap — any
// caller handing it a stale or hand-built plan. Hand it the sick
// Halfling explicitly: it must refuse to tap and must mint nothing.
func TestMaterializePlanRefusesSummoningSickManaCreature(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	id := pushIntrinsicPermanent(g, p, "Delighted Halfling", "Creature — Halfling Citizen", halflingAbilities(), nil)
	setSummonedThisTurn(g, id, true)

	g.WithWriteLock(func() {
		g.materializePlanLocked(p, []uuid.UUID{id}, costFor(t, "{1}"))
	})
	if cardTapped(g, id) {
		t.Error("the executor tapped a creature that entered this turn")
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool = %+v, want empty", p.ManaPool)
	}
}

// End-to-end through the cast path, which is how the reporter hit it:
// AutoTap must not fund a spell off a creature that just landed.
func TestCastAutoTapWontFundOffSummoningSickCreature(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	halfling := pushIntrinsicPermanent(g, p, "Delighted Halfling", "Creature — Halfling Citizen", halflingAbilities(), nil)
	setSummonedThisTurn(g, halfling, true)
	spell := pushTypedCardToHandWithCost(p, "Ornithopter", "Artifact Creature — Thopter", "{1}")

	err := g.CastSpell(p.ID, spell, CastSpellParams{Strict: true, AutoTap: true})
	if err == nil {
		t.Fatal("auto-tap funded a cast off a summoning-sick mana creature")
	}
	if cardTapped(g, halfling) {
		t.Error("the refused cast tapped the Halfling anyway")
	}
}
