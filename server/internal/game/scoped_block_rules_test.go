package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_block_rules_test.go pins ADR 0041 phase 3 tier 3b-2 for the
// block-rule kinds (#1497): cantBeBlockedExceptBy and
// limitBlockersPerDefender are ScopedEffect records, read by
// forEachScopedBlockRuleLocked rather than the layer pass. The card
// behaviour (Gingerbrute, Departed Deckhand, Mirri) is pinned in
// cards/effects and in declare_blockers_test.go /
// conditional_combat_limits_test.go here; this file pins the engine
// adapter directly.

func TestBlockRuleModProblemCatchesBadParameters(t *testing.T) {
	bad := []Mod{
		{Kind: ModCantBeBlockedExceptBy, Keywords: []string{"haste"}},    // no Text
		{Kind: ModCantBeBlockedExceptBy, Text: "creatures with haste"},   // no keyword or subtype
		{Kind: ModLimitBlockersPerDefender, Amount: 0},                   // amount below 1
		{Kind: ModCantBeBlockedByPlayer, Text: "creatures Bob controls"}, // no player
		{Kind: ModCantBeBlockedByPlayer, Player: uuid.New()},             // no text
		{Kind: ModCantBeBlockedByPower, Amount: 2},                       // no text
	}
	for _, m := range bad {
		if blockRuleModProblem(m) == "" {
			t.Errorf("%+v: want a problem, got none", m)
		}
	}
	ok := []Mod{
		{Kind: ModCantBeBlockedExceptBy, Keywords: []string{"haste"}, Text: "creatures with haste"},
		{Kind: ModCantBeBlockedExceptBy, Subtypes: []string{"Spirit"}, Text: "Spirits"},
		{Kind: ModLimitBlockersPerDefender, Amount: 1},
		{Kind: ModCantBeBlockedByPlayer, Player: uuid.New(), Text: "creatures Bob controls"},
		{Kind: ModCantBeBlockedByPower, Amount: 2, Text: "creatures with power 2 or less"},
		{Kind: ModCantBeBlockedByPower, Amount: 0, Text: "creatures with power 0 or less"},
	}
	for _, m := range ok {
		if p := blockRuleModProblem(m); p != "" {
			t.Errorf("%+v: unexpected problem %q", m, p)
		}
	}
}

// A block-rule mod is not a layer operation: the layer adapter builds
// nothing for it.
func TestTheLayerAdapterSkipsBlockRuleMods(t *testing.T) {
	recs := []ScopedEffect{{
		Affected: []AffectedObject{{ID: uuid.New()}},
		Mods:     []Mod{{Kind: ModCantBeBlockedExceptBy, Keywords: []string{"haste"}, Text: "creatures with haste"}},
		Seq:      1,
	}}
	if got := adaptScopedEffects(recs); len(got) != 0 {
		t.Errorf("the layer adapter built %d continuous effects from a block-rule record", len(got))
	}
}

// cantBeBlockedExceptBy: the pinned attacker is refused by a blocker
// with neither the keyword nor the subtype, and allowed by either —
// including a changeling for the subtype half.
func TestCantBeBlockedExceptByAllowsKeywordOrSubtype(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Attacker", 2, 2)
	plain := pushCombatant(t, g, opp, "Plain Blocker", 2, 2)
	hasty := pushCombatant(t, g, opp, "Hasty Blocker", 2, 2, "haste")
	shifter := pushCombatant(t, g, opp, "Changeling", 1, 1, KeywordChangeling)

	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(attacker),
			[]Mod{CantBeBlockedExceptByMod([]string{"haste"}, []string{"Spirit"}, "creatures with haste or Spirits")},
			IndefiniteDuration(), "test") {
			t.Fatal("setup: registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})

	refusal := func(blocker uuid.UUID) BlockRefusal {
		var r BlockRefusal
		g.WithWriteLock(func() {
			r = g.blockRuleRefusalLocked(findCard(g, attacker), findCard(g, blocker))
		})
		return r
	}
	if r := refusal(plain); r.Legal() {
		t.Error("a blocker with neither the keyword nor the subtype was allowed")
	} else if r.Reason != BlockReasonCantBeBlockedExceptBy || r.Label != "creatures with haste or Spirits" {
		t.Errorf("refusal = %+v", r)
	}
	if r := refusal(hasty); !r.Legal() {
		t.Errorf("a hasty blocker was refused: %+v", r)
	}
	if r := refusal(shifter); !r.Legal() {
		t.Errorf("a changeling was refused: %+v", r)
	}
}

// limitBlockersPerDefender: the record's controller's own creatures
// never count, and an opponent's creature counts toward the bound.
func TestLimitBlockersPerDefenderCountsOpponentsOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushCombatant(t, g, me, "My Blocker", 2, 2)
	theirs := pushCombatant(t, g, opp, "Their Blocker", 2, 2)

	g.WithWriteLock(func() {
		if !g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeOpponentsCreatures, me.ID,
			[]Mod{LimitBlockersPerDefenderMod(1)}, IndefiniteDuration(), "test limit") {
			t.Fatal("setup: registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})

	limitFor := func(blocker uuid.UUID) int {
		n := 0
		g.WithWriteLock(func() {
			g.forEachScopedBlockRuleLocked(func(r BlockRule, source *Card) bool {
				if r.Limit != nil {
					n = r.Limit(g, findCard(g, blocker), source)
				}
				return true
			})
		})
		return n
	}
	if got := limitFor(mine); got != 0 {
		t.Errorf("the record's own controller's creature counts: limit %d, want 0", got)
	}
	if got := limitFor(theirs); got != 1 {
		t.Errorf("an opponent's creature: limit %d, want 1", got)
	}
}

// cantBeBlockedByPower (#2600): a blocker of power N or less is refused,
// N+1 is allowed, an unpinned attacker is not covered, and the blocker's
// power is read live: a pump made after the effect moves it out of the
// barred set, a shrink moves it in.
func TestCantBeBlockedByPowerReadsTheBlockersPowerLive(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Attacker", 3, 3)
	other := pushCombatant(t, g, me, "Other Attacker", 3, 3)
	small := pushCombatant(t, g, opp, "Power Two", 2, 2)
	big := pushCombatant(t, g, opp, "Power Three", 3, 3)

	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(attacker),
			[]Mod{CantBeBlockedByPowerMod(2, "creatures with power 2 or less")},
			IndefiniteDuration(), "test") {
			t.Fatal("setup: registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	refusal := func(att, blocker uuid.UUID) BlockRefusal {
		var r BlockRefusal
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			r = g.blockRuleRefusalLocked(findCard(g, att), findCard(g, blocker))
		})
		return r
	}
	if r := refusal(attacker, small); r.Legal() {
		t.Error("a blocker of power N was allowed")
	} else if r.Reason != BlockReasonCantBeBlockedBy || r.Label != "creatures with power 2 or less" {
		t.Errorf("refusal = %+v", r)
	}
	if r := refusal(attacker, big); !r.Legal() {
		t.Errorf("a blocker of power N+1 was refused: %+v", r)
	}
	if r := refusal(other, small); !r.Legal() {
		t.Errorf("an attacker the effect is not pinned to was covered: %+v", r)
	}

	// A pump on the blocker after the effect was applied is read live.
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(small),
			[]Mod{ModifyPTMod(1, 0)}, IndefiniteDuration(), "pump")
	})
	if r := refusal(attacker, small); !r.Legal() {
		t.Errorf("a blocker pumped past N after the effect was still refused: %+v", r)
	}
	// And a shrink pulls a bigger creature in.
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(big),
			[]Mod{ModifyPTMod(-1, 0)}, IndefiniteDuration(), "shrink")
	})
	if r := refusal(attacker, big); r.Legal() {
		t.Error("a blocker shrunk to power N after the effect was still allowed")
	}
}
