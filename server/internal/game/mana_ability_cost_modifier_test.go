package game

import "testing"

// mana_ability_cost_modifier_test.go — #1191: the CR 601.2f cost pass
// reaches a CR 605 mana ability's own activation cost, not only a
// CR 602 ability's. #1193 built CostQuery.Ability and
// AbilityCostSubject.Mana for exactly this and left the field always
// false; this file is the first reader to set it.
//
// The probes use a synthetic mana ability with a plain GENERIC mana
// component ("{2}, {T}: Add {C}{C}{C}") rather than Loot, the
// Pathfinder's printed "{G}, {T}": a reduction only ever spends
// generic mana (CR 601.2f), so a colored requirement like Loot's {G}
// can never be the thing that proves a discount landed. A synthetic
// Signet-shaped ability isolates the property under test from that
// unrelated fact about the one printed exhaust mana ability.

const manaCostAbilityLabel = "{2}, {T}: Add {C}{C}{C}"

// manaCostAbility is the Signet-shaped probe: TapCost plus a real
// mana component in its OWN activation cost.
func manaCostAbility() ManaAbilityShape {
	return ManaAbilityShape{
		TapCost:  true,
		ManaCost: "{2}",
		Produced: "{C}{C}{C}",
		Label:    manaCostAbilityLabel,
	}
}

// oneGenericManaCostAbility costs {1} rather than {2}, so a {2}
// reduction (floored at zero generic) empties it out completely —
// the shape the "planned" half of #1191 needs, since the auto-tapper
// only ever plans a mana component that prices to nothing.
func oneGenericManaCostAbility() ManaAbilityShape {
	return ManaAbilityShape{
		TapCost:  true,
		ManaCost: "{1}",
		Produced: "{C}{C}{C}",
		Label:    "{1}, {T}: Add {C}{C}{C}",
	}
}

// boomScholarShapedDiscount is Boom Scholar's own clause, reused
// verbatim from exhaust_readers_test.go's shape but WITHOUT the
// exhaust predicate — these probes are ordinary (non-exhaust) mana
// abilities, so the modifier here reads only "another permanent you
// control", which is the half of Boom Scholar's AppliesTo that is not
// specific to the exhaust keyword.
func boomScholarShapedDiscount() CostModifier {
	return CostModifier{
		Kind:        CostReduction,
		Activations: true,
		Label:       "Exhaust abilities of other permanents you control cost {2} less to activate.",
		AppliesTo: func(q CostQuery) bool {
			return q.Ability != nil && q.Card.Controller == q.Source.Controller &&
				q.Card.InstanceID != q.Source.InstanceID
		},
		Amount: func(CostQuery) int { return 2 },
	}
}

// TestAManaAbilityCostModifierReadsTheAbility is #1191's charging
// half, mirroring TestAnActivationCostModifierReadsTheAbility for the
// CR 605 path: ManaAbilityManaCostForEffect is the first function to
// ask CostQuery.Ability.Mana's question, and this is the accessor-level
// proof that it prices a mana ability's own cost at all.
func TestAManaAbilityCostModifierReadsTheAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-price", manaCostAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-price")

	var cost ParsedCost
	g.WithWriteLock(func() {
		card := *g.findCardByIDLocked(src)
		cost, _ = g.ManaAbilityManaCostForEffect(me.ID, card, ManaAbilitiesForCard(card)[0])
	})
	if cost.Generic != 2 {
		t.Fatalf("with no discounter on the board: {%d}, want the printed {2}", cost.Generic)
	}

	costModifierCatalog(t, "probe-mana-discount", boomScholarShapedDiscount())
	pushManaExhaustSource(g, me, "probe-mana-discount")
	g.WithWriteLock(func() {
		card := *g.findCardByIDLocked(src)
		cost, _ = g.ManaAbilityManaCostForEffect(me.ID, card, ManaAbilitiesForCard(card)[0])
	})
	if cost.Generic != 0 {
		t.Errorf("the mana ability costs {%d}, want {0} — {2} reduced by {2}", cost.Generic)
	}
}

// TestANilAppliesToCostModifierDoesNotPriceAManaAbility is the
// partition, mirroring TestACastModifierDoesNotPriceAnActivation for
// the CR 605 path: a CAST modifier with a nil AppliesTo (Sphere of
// Resistance's "every spell") must not reach a mana ability's cost
// now that CostQuery.Ability is non-nil for one — the SAME
// Activations-against-Ability-nilness partition activeCostModifiersLocked
// already runs, unchanged by this file, but the mana wiring is new
// enough that the property is worth pinning here rather than only
// trusting the CR 602 test to cover it.
func TestANilAppliesToCostModifierDoesNotPriceAManaAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-tax", manaCostAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-tax")
	costModifierCatalog(t, "probe-mana-sphere", CostModifier{
		Kind:   CostIncrease,
		Label:  "Spells cost {1} more to cast.",
		Amount: func(CostQuery) int { return 1 },
	})
	pushManaExhaustSource(g, me, "probe-mana-sphere")

	var cost ParsedCost
	g.WithWriteLock(func() {
		card := *g.findCardByIDLocked(src)
		cost, _ = g.ManaAbilityManaCostForEffect(me.ID, card, ManaAbilitiesForCard(card)[0])
	})
	if cost.Generic != 2 {
		t.Errorf("the mana ability costs {%d}, want the printed {2} — a nil-AppliesTo CAST modifier prices casts, not mana abilities", cost.Generic)
	}
}

// TestADiscountedManaAbilityCostIsWhatActivationActuallyPays is
// #1191's other half of "charged": the pass has to be on the PAYING
// path (ActivateManaAbility), not only in an accessor, or the engine
// and the legal-move enumerator would agree with each other and both
// be wrong (#544). An empty pool proves it: paying the printed {1}
// would fail outright.
func TestADiscountedManaAbilityCostIsWhatActivationActuallyPays(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-pay", oneGenericManaCostAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-pay")
	costModifierCatalog(t, "probe-mana-pay-discount", boomScholarShapedDiscount())
	pushManaExhaustSource(g, me, "probe-mana-pay-discount")

	me.ManaPool = nil // empty pool: paying the printed {1} would fail
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activation with an empty pool: %v (want no error — the discount took the printed {1} to {0})", err)
	}
	if len(me.ManaPool) != 3 {
		t.Errorf("pool = %v, want the produced {C}{C}{C} — the {1} discounted to {0} spent nothing", me.ManaPool)
	}
}

// TestADiscountedManaAbilityCostIsPlannedByTheAutoTapper is #1191's
// "planned" half: autoTapAbilityFor refuses any mana ability with a
// ManaCost component, unconditionally, before this fix — the file
// header's standing MANA-cost bullet. A component a modifier prices
// to nothing leaves the planner no decision to make, so the source
// becomes plannable exactly as if it had never printed the
// component; a component still owing something after the discount
// keeps the standing refusal (unchanged, and not exercised here —
// see TestTapOthersManaAbilityIsNeverAutoTapped and the ordinary
// Signet case for that side).
func TestADiscountedManaAbilityCostIsPlannedByTheAutoTapper(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustManaCatalog(t, "probe-mana-plan", oneGenericManaCostAbility())
	src := pushManaExhaustSource(g, me, "probe-mana-plan")

	// Non-vacuity: undiscounted, the source is not plannable at all.
	g.mu.Lock()
	before := g.autoTapAbilityFor(me.ID, *g.findCardByIDLocked(src), ManaAbilitiesForCard(*g.findCardByIDLocked(src)))
	g.mu.Unlock()
	if before != nil {
		t.Fatalf("test setup: a printed {1} mana cost should not be plannable before any discount, got %+v", before)
	}

	costModifierCatalog(t, "probe-mana-plan-discount", boomScholarShapedDiscount())
	pushManaExhaustSource(g, me, "probe-mana-plan-discount")

	cost := costFor(t, "{C}{C}{C}")
	g.WithWriteLock(func() {
		plan, ok := g.autoTapLocked(me.ID, cost, 0, nil)
		if !ok || len(plan) != 1 || plan[0].CardID != src {
			t.Fatalf("plan = %+v (ok=%v), want the discounted source alone", plan, ok)
		}
		g.materializePlanLocked(me, plan, cost)
	})
	if len(me.ManaPool) != 3 {
		t.Fatalf("pool = %v, want the planned {C}{C}{C} — the executor must not have paid the {1} it never priced", me.ManaPool)
	}
	if !tappedForTest(g, src) {
		t.Errorf("the plan did not tap the source")
	}
}
