package game

import (
	"testing"

	"github.com/google/uuid"
)

// autotap_sacrifice_test.go — #1215. The auto-tapper used to refuse
// EVERY mana ability with a sacrifice component, which folded two
// unlike clauses into one exclusion: "sacrifice OTHER permanents"
// (Ashnod's Altar) names victims and is a decision the planner may not
// make, while "sacrifice this" (a Treasure, a Lotus Petal) names
// nothing and needs no decision at all. A board of Treasures read as
// unpayable to the preview, the strict cast gate and every bot.
//
// These tests hold the split in both directions, hold the executor to
// ActivateManaAbility's component order, and hold the LAST-RESORT
// ordering that keeps the planner from cracking a Treasure for a pip a
// land could have paid.

// treasureManaAbility is the printed Treasure: "{T}, Sacrifice this
// artifact: Add one mana of any color."
func treasureManaAbility() []ManaAbilityShape {
	return []ManaAbilityShape{{
		TapCost:       true,
		SacrificeCost: true,
		Produced:      "{W|U|B|R|G}",
		Label:         "{T}, Sacrifice: Add one mana of any color",
	}}
}

// pushTreasures mints n Treasure tokens under `owner` and returns
// their IDs.
func pushTreasures(g *Game, owner *Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, pushIntrinsicPermanent(g, owner,
			"Treasure", "Token Artifact — Treasure", treasureManaAbility(), nil))
	}
	return out
}

func containsID(ids []uuid.UUID, want uuid.UUID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

// The headline: N Treasures and nothing else pay an N-cost spell. This
// planned nothing at all before the split — `ok` came back false and
// the cast reported the whole cost as missing.
func TestAutoTapPlansABoardOfTreasures(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	treasures := pushTreasures(g, me, 3)

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{3}"), 0)
	if !ok {
		t.Fatalf("three Treasures should pay {3}")
	}
	if len(plan) != 3 {
		t.Fatalf("plan = %v, want all three Treasures", plan)
	}
	for _, id := range treasures {
		if !containsID(plan, id) {
			t.Errorf("Treasure %v missing from the plan %v", id, plan)
		}
	}
}

// A Treasure is a five-colour source, so it pays a coloured pip too.
func TestAutoTapPlansATreasureForAColouredPip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	treasure := pushTreasures(g, me, 1)[0]

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{U}"), 0)
	if !ok || len(plan) != 1 || plan[0] != treasure {
		t.Fatalf("one Treasure for {U}: ok=%v plan=%v, want the Treasure", ok, plan)
	}
}

// The other half of the split stays shut. Ashnod's Altar asks which
// creature dies, and the planner answers no questions — the card is
// still hand-activated from the permanent's ability menu.
func TestAutoTapStillRefusesSacrificeAnother(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushIntrinsicPermanent(g, me, "Ashnod's Altar", "Artifact", altarAbility(), nil)
	pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)

	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0); ok {
		t.Fatalf("the Altar was planned: %v — sacrificing a named creature is not the planner's decision", plan)
	}
	// And a Treasure that ALSO ate something else would be the same
	// decision, so the compound shape stays out with it.
	both := treasureManaAbility()
	both[0].SacrificeOther = altarAbility()[0].SacrificeOther
	pushIntrinsicPermanent(g, me, "Greedy Treasure", "Token Artifact — Treasure", both, nil)
	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0); ok {
		t.Fatalf("a sacrifice-self AND sacrifice-another ability was planned: %v", plan)
	}
}

// The last-resort tier. A Treasure is a five-colour source, so the
// any-colour generic tier would have recruited it AHEAD of the basic
// land beside it — which spends a resource the player never agreed to
// spend on a pip the land pays for free and untaps from.
func TestAutoTapPrefersLandsOverTreasuresForGeneric(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mountains := []uuid.UUID{
		pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", ""),
		pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", ""),
	}
	treasures := pushTreasures(g, me, 2)

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{2}"), 0)
	if !ok || len(plan) != 2 {
		t.Fatalf("two Mountains and two Treasures for {2}: ok=%v plan=%v", ok, plan)
	}
	for _, id := range treasures {
		if containsID(plan, id) {
			t.Errorf("plan %v cracked a Treasure for generic mana two Mountains could pay", plan)
		}
	}
	for _, id := range mountains {
		if !containsID(plan, id) {
			t.Errorf("plan %v skipped a Mountain", plan)
		}
	}
}

// Same instinct on the coloured pass: the Mountain pays the {R} and
// the Treasure stays on the battlefield, even though both can.
func TestAutoTapPrefersALandOverATreasureForAColouredPip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mountain := pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	treasure := pushTreasures(g, me, 1)[0]

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{R}"), 0)
	if !ok || len(plan) != 1 {
		t.Fatalf("Mountain + Treasure for {R}: ok=%v plan=%v", ok, plan)
	}
	if plan[0] != mountain {
		t.Errorf("plan = %v, want the Mountain, not the Treasure %v", plan, treasure)
	}
}

// …and the Treasure is still reached when the land alone is short. The
// last-resort tier is an ORDER, not a second exclusion.
func TestAutoTapReachesTheTreasureWhenTheLandsAreShort(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mountain := pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	treasure := pushTreasures(g, me, 1)[0]

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{2}"), 0)
	if !ok || len(plan) != 2 {
		t.Fatalf("Mountain + Treasure for {2}: ok=%v plan=%v", ok, plan)
	}
	if !containsID(plan, mountain) || !containsID(plan, treasure) {
		t.Errorf("plan = %v, want both sources", plan)
	}
}

// The executor's half: the plan is materialised, the Treasures are
// tapped and then sacrificed, and the mana lands in the pool. The
// component order is ActivateManaAbility's — tap, then sacrifice, then
// mana — so a "whenever this becomes tapped" watcher sees the tap.
func TestMaterializePlanCracksTheTreasuresItTaps(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	treasures := pushTreasures(g, me, 2)

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{2}"), 0)
	if !ok {
		t.Fatalf("two Treasures should pay {2}")
	}
	planned := make(tapPlan, 0, len(plan))
	for _, id := range plan {
		planned = append(planned, plannedTap{CardID: id})
	}
	g.WithWriteLock(func() { g.materializePlanLocked(me, planned, costFor(t, "{2}")) })

	for _, id := range treasures {
		if g.Battlefield.Contains(id) {
			t.Errorf("Treasure %v survived the payment", id)
		}
		if !me.Graveyard.Contains(id) {
			t.Errorf("Treasure %v never reached the graveyard", id)
		}
		if !hasEvent(g, EventSacrifice, id) {
			t.Errorf("no EventSacrifice for Treasure %v", id)
		}
		if !hasEvent(g, EventTapCard, id) {
			t.Errorf("no EventTapCard for Treasure %v — the tap must precede the sacrifice", id)
		}
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %+v, want two mana", me.ManaPool)
	}
}

// A plan can arrive stale. A Treasure that changed hands between the
// plan and the payment is not this player's to tap OR to sacrifice
// (CR 701.21a), so the executor drops it and mints nothing — rather
// than eating an opponent's permanent.
func TestMaterializePlanRefusesASourceThatChangedHands(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	treasure := pushTreasures(g, me, 1)[0]
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == treasure {
			g.Battlefield.Cards[i].Controller = opp.ID
		}
	}

	g.WithWriteLock(func() {
		g.materializePlanLocked(me, tapPlan{{CardID: treasure}}, costFor(t, "{1}"))
	})
	if !g.Battlefield.Contains(treasure) {
		t.Error("the executor sacrificed a permanent the payer no longer controls")
	}
	if cardTapped(g, treasure) {
		t.Error("the executor tapped a permanent the payer no longer controls")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %+v, want empty", me.ManaPool)
	}
}

// End to end, which is the shape the bug was reported in: three
// Treasures, no lands, and a {3} spell that the strict gate refused
// because the planner had nothing to offer it.
func TestCastAutoTapPaysAThreeCostSpellOffThreeTreasures(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	treasures := pushTreasures(g, me, 3)
	spell := pushTypedCardToHandWithCost(me, "Icy Manipulator", "Artifact", "{3}")

	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tap cast off three Treasures: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Error("the spell never reached the stack")
	}
	for _, id := range treasures {
		if g.Battlefield.Contains(id) {
			t.Errorf("Treasure %v survived a cast it paid for", id)
		}
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %+v, want the three mana spent on the cast", me.ManaPool)
	}
}

// And the same cast is still refused when the board is one Treasure
// short — the split must not turn "can't pay" into "pay anyway".
func TestCastAutoTapStillRefusesWhenTheTreasuresAreShort(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	treasures := pushTreasures(g, me, 2)
	spell := pushTypedCardToHandWithCost(me, "Icy Manipulator", "Artifact", "{3}")

	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("two Treasures funded a {3} spell")
	}
	for _, id := range treasures {
		if !g.Battlefield.Contains(id) {
			t.Errorf("the refused cast cracked Treasure %v anyway", id)
		}
	}
}

// --- #1212 × #1215: the wish and the last-resort tier ---------------

// The two features meet on exactly one kind of source. A Treasure is
// both the commonest source a card asks for BY NAME ("if mana from a
// Treasure was spent to cast it") and a source the payment destroys,
// so the last-resort tier and #1212's source wish pull in opposite
// directions on the same permanent.
//
// The card's own text wins. Board: two Mountains and a Treasure, cost
// {1}{R} — one Mountain pays the pip and the generic is a straight
// choice between the second Mountain and the Treasure. With no wish
// the tier keeps the Treasure whole; with a Treasure wish the planner
// cracks it, because a Hired Hexblade paid for with the Mountain is a
// Hexblade whose printed text does nothing.
func TestTheSourceWishOutranksTheSacrificeTier(t *testing.T) {
	build := func() (*Game, *Player, uuid.UUID) {
		g := newActiveGame(t)
		me := g.Seats[0]
		pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
		pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
		return g, me, pushTreasures(g, me, 1)[0]
	}
	var (
		plan tapPlan
		ok   bool
	)

	// No wish: the tier holds and both Mountains pay.
	g, me, treasure := build()
	g.ReadSnapshot(func() {
		plan, ok = g.autoTapPreferringLocked(me.ID, costFor(t, "{1}{R}"), 0, nil, 0)
	})
	if !ok || len(plan) != 2 {
		t.Fatalf("no wish: ok=%v plan=%v, want two taps", ok, plan)
	}
	if plan.hasCard(treasure) {
		t.Errorf("no wish: plan %v cracked a Treasure two Mountains could pay for", plan.cardIDs())
	}

	// A Treasure wish: the planner reaches past the tier for it.
	g, me, treasure = build()
	g.ReadSnapshot(func() {
		plan, ok = g.autoTapPreferringLocked(me.ID, costFor(t, "{1}{R}"), 0, nil, ManaSourceTreasure)
	})
	if !ok || len(plan) != 2 {
		t.Fatalf("Treasure wish: ok=%v plan=%v, want two taps", ok, plan)
	}
	if !plan.hasCard(treasure) {
		t.Errorf("Treasure wish: plan %v left the Treasure alone — the wish must outrank the last-resort "+
			"tier, or it is inert for the one family it was written for", plan.cardIDs())
	}
}

// The same claim against the GENERIC RECRUITER alone. The test above
// is satisfied by the coloured pass — a Treasure is a five-colour
// source, so a wished one pays the {R} and never reaches
// orderUnusedByGenericPreference — and the two comparators are
// separate code that can be reordered separately. A cost with NO
// coloured pip has only the recruiter in it, so this is the one that
// fails if that comparator's keys are swapped back.
func TestTheSourceWishOutranksTheSacrificeTierWhenRecruitingGeneric(t *testing.T) {
	build := func() (*Game, *Player, uuid.UUID) {
		g := newActiveGame(t)
		me := g.Seats[0]
		pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
		pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
		return g, me, pushTreasures(g, me, 1)[0]
	}
	var (
		plan tapPlan
		ok   bool
	)

	g, me, treasure := build()
	g.ReadSnapshot(func() {
		plan, ok = g.autoTapPreferringLocked(me.ID, costFor(t, "{2}"), 0, nil, 0)
	})
	if !ok || len(plan) != 2 {
		t.Fatalf("no wish, {2}: ok=%v plan=%v", ok, plan)
	}
	if plan.hasCard(treasure) {
		t.Errorf("no wish, {2}: plan %v cracked a Treasure two Mountains could pay", plan.cardIDs())
	}

	g, me, treasure = build()
	g.ReadSnapshot(func() {
		plan, ok = g.autoTapPreferringLocked(me.ID, costFor(t, "{2}"), 0, nil, ManaSourceTreasure)
	})
	if !ok || len(plan) != 2 {
		t.Fatalf("Treasure wish, {2}: ok=%v plan=%v", ok, plan)
	}
	if !plan.hasCard(treasure) {
		t.Errorf("Treasure wish, {2}: plan %v left the Treasure alone — the generic recruiter must read "+
			"the wish before the last-resort tier too", plan.cardIDs())
	}
}

// The wish reaches the COLOURED pass as well as the generic
// recruiter, because it is the first key of both comparators: a
// Treasure is a five-colour source, and a wished one pays the {R}
// ahead of the Mountain that is otherwise both cheaper and more
// restrictive.
func TestTheSourceWishReachesTheColouredPass(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	treasure := pushTreasures(g, me, 1)[0]

	var (
		plan tapPlan
		ok   bool
	)
	g.ReadSnapshot(func() {
		plan, ok = g.autoTapPreferringLocked(me.ID, costFor(t, "{R}"), 0, nil, ManaSourceTreasure)
	})
	if !ok || len(plan) != 1 {
		t.Fatalf("Treasure wish for {R}: ok=%v plan=%v", ok, plan)
	}
	if plan[0].CardID != treasure {
		t.Errorf("plan = %v, want the wished Treasure %v", plan.cardIDs(), treasure)
	}
}

// …and it stays a HINT. A wish nothing on the board satisfies changes
// nothing: the candidate set is identical with and without one, so a
// cost that was payable stays payable and the last-resort tier still
// governs everything the wish did not name.
func TestAnUnmatchedWishLeavesTheTierInCharge(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	treasure := pushTreasures(g, me, 1)[0]

	var (
		plan tapPlan
		ok   bool
	)
	g.ReadSnapshot(func() {
		// Nothing on this board is a snow source.
		plan, ok = g.autoTapPreferringLocked(me.ID, costFor(t, "{1}{R}"), 0, nil, ManaSourceSnow)
	})
	if !ok || len(plan) != 2 {
		t.Fatalf("unmatched wish: ok=%v plan=%v", ok, plan)
	}
	if plan.hasCard(treasure) {
		t.Errorf("unmatched wish: plan %v cracked a Treasure — an unmatched wish must leave the tier in charge",
			plan.cardIDs())
	}
}
