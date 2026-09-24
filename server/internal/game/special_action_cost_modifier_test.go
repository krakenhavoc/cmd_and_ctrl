package game

import "testing"

// special_action_cost_modifier_test.go — #1319: a CR 116.2 special
// action's cost is now priced through the CR 601.2f pass, exactly as
// a cast (S28) and an activation (#1184) are. What is pinned here is
// the new partition (CostModifier.SpecialActions), in both
// directions, and the per-turn tally (Game.ForetoldThisTurn) Ranar
// the Ever-Watchful's "the first card you foretell each turn costs
// {0} to foretell" reads.

const forestellCostModifierCardOracle = "test-foretell-discount-card"
const forestellCostModifierSourceOracle = "test-foretell-discount-source"

// ranarStyleDiscount is a CostModifier in Ranar's shape.
func ranarStyleDiscount() CostModifier {
	return CostModifier{
		Kind:           CostReduction,
		SpecialActions: true,
		Label:          "The first card you foretell each turn costs {0} to foretell.",
		AppliesTo: func(q CostQuery) bool {
			return q.SpecialAction != nil && q.SpecialAction.Kind == SpecialActionForetell &&
				q.Game.ForetoldCountThisTurn(q.Controller) == 0
		},
		Amount: fixed(2),
	}
}

func withForetellDiscount(t *testing.T) {
	t.Helper()
	withCatalogSpecialActions(t, func(id string) []SpecialAction {
		if id != forestellCostModifierCardOracle {
			return nil
		}
		return []SpecialAction{{Kind: SpecialActionForetell, Cost: ForetellExileCost, CastCost: "{1}{U}", Label: "Foretell {2}"}}
	})
	withCatalogCostModifiers(t, modifiersFor(forestellCostModifierSourceOracle, ranarStyleDiscount()))
}

// TestFirstForetellEachTurnIsFreeButTheSecondIsNot is the end-to-end
// shape, on the PAYING path (TestTheDiscountIsWhatTheActivationActuallyPays's
// discipline, one verb over): the discount reaches the first foretell
// of the turn and stops reaching the second, because
// Game.ForetoldThisTurn is bumped once the first one actually lands
// — "the FIRST card", not "the first N".
func TestFirstForetellEachTurnIsFreeButTheSecondIsNot(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	withForetellDiscount(t)
	modifierSource(t, g, me, "Ranar the Ever-Watchful", forestellCostModifierSourceOracle)

	first := seedHandCard(me, "Saw It Coming", forestellCostModifierCardOracle, "Instant", "{1}{U}")
	second := seedHandCard(me, "Behold the Multiverse", forestellCostModifierCardOracle, "Instant", "{2}{U}")

	if got := g.ForetoldCountThisTurn(me.ID); got != 0 {
		t.Fatalf("ForetoldCountThisTurn before any foretell: got %d, want 0", got)
	}

	// STRICT mode with an EMPTY pool: this is the whole assertion.
	// Permissive mode would wave a {2} through with a cost warning
	// and the test would pass whether or not the discount actually
	// priced anything.
	if err := g.PerformSpecialAction(me.ID, first.InstanceID, SpecialActionForetell, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("first foretell (should be free): %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool after the free foretell: got %d tokens, want 0", len(me.ManaPool))
	}
	if got := g.ForetoldCountThisTurn(me.ID); got != 1 {
		t.Fatalf("ForetoldCountThisTurn after one foretell: got %d, want 1", got)
	}

	// The second foretell this turn costs the full printed {2}.
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, second.InstanceID, SpecialActionForetell, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("second foretell (should cost {2}, and the pool can pay it): %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool after the second foretell: got %d tokens, want 0 — the printed {2} should have been charged", len(me.ManaPool))
	}
	if got := g.ForetoldCountThisTurn(me.ID); got != 2 {
		t.Errorf("ForetoldCountThisTurn after two foretells: got %d, want 2", got)
	}
}

// TestTheForetellTallyResetsNextTurn: "each turn", not "each game".
func TestTheForetellTallyResetsNextTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	withForetellDiscount(t)
	modifierSource(t, g, me, "Ranar the Ever-Watchful", forestellCostModifierSourceOracle)

	first := seedHandCard(me, "Saw It Coming", forestellCostModifierCardOracle, "Instant", "{1}{U}")
	if err := g.PerformSpecialAction(me.ID, first.InstanceID, SpecialActionForetell, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("first foretell: %v", err)
	}
	if got := g.ForetoldCountThisTurn(me.ID); got != 1 {
		t.Fatalf("ForetoldCountThisTurn after one foretell: got %d, want 1", got)
	}

	// Walk the cursor all the way around to precombat main again,
	// which lands on a later turn (onTurnBeganLocked has run at least
	// once) whatever the seat count.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceTo(t, g, StepPrecombatMain)

	if got := g.ForetoldCountThisTurn(me.ID); got != 0 {
		t.Errorf("ForetoldCountThisTurn on a new turn: got %d, want 0 — the tally should have cleared", got)
	}
}

// TestASpecialActionCostModifierDoesNotPriceAnOrdinaryCast is #1319's
// partition, the direction TestACastModifierDoesNotPriceAnActivation
// does not cover: a modifier written for a special action must never
// reach an ordinary spell cast.
func TestASpecialActionCostModifierDoesNotPriceAnOrdinaryCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, modifiersFor(forestellCostModifierSourceOracle, ranarStyleDiscount()))
	modifierSource(t, g, me, "Ranar the Ever-Watchful", forestellCostModifierSourceOracle)

	spell := spellInHand(t, g, me, "Lightning Bolt", "Instant", "{R}")
	cost := priceOf(t, g, me, spell, CastSpellParams{})
	if cost.String() != "{R}" {
		t.Errorf("a special-action cost modifier reached an ordinary cast: priced at %s, want {R}", cost.String())
	}
}

// TestACastModifierDoesNotMakeAForetellFree is the same partition,
// the other direction: an ordinary "spells cost {N} less" must never
// reach a special action.
func TestACastModifierDoesNotMakeAForetellFree(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	withCatalogSpecialActions(t, func(id string) []SpecialAction {
		if id != forestellCostModifierCardOracle {
			return nil
		}
		return []SpecialAction{{Kind: SpecialActionForetell, Cost: ForetellExileCost, CastCost: "{1}{U}", Label: "Foretell {2}"}}
	})
	withCatalogCostModifiers(t, modifiersFor(forestellCostModifierSourceOracle, CostModifier{
		Kind:   CostReduction,
		Label:  "Spells you cast cost {2} less to cast.",
		Amount: fixed(2),
	}))
	modifierSource(t, g, me, "Goblin Electromancer", forestellCostModifierSourceOracle)
	card := seedHandCard(me, "Saw It Coming", forestellCostModifierCardOracle, "Instant", "{1}{U}")

	// STRICT with an empty pool: a spell-shaped reduction reaching
	// the special action would pay for nothing and this would pass.
	if err := g.PerformSpecialAction(me.ID, card.InstanceID, SpecialActionForetell, SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("a spell cost modifier discounted a foretell to where an empty pool could pay it")
	}
}
