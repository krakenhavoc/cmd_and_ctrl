package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// block_rules_test.go — #750, CR 509.1b: the block restrictions that
// carry a PARAMETER ("can't be blocked except by Walls", "can't be
// blocked by creatures with power 2 or less", "can't be blocked by
// more than one creature").
//
// Two shapes, two enforcement points, and the split is the thing these
// tests pin:
//
//   - A PAIR rule is per-pair, so it is slot 4 of
//     BlockPairRefusalLocked and DeclareBlocker refuses the block
//     outright, with a reason the player can read.
//   - A COUNT rule is a bound on a whole declaration, so it cannot be
//     judged pair by pair. It is judged at the declaration's lock-in,
//     where the declaration is complete (CR 509.1), and an illegal set
//     is reverted there — before any EventBlock is emitted, so no
//     "whenever this creature blocks" trigger ever sees it.

// blockRuleOracle is the catalog key the stubbed hook answers for. A
// test stamps it onto the permanent that is supposed to carry the
// rule, so every other permanent on the board carries none.
const blockRuleOracle = "block-rule-oracle-test"

// stubCatalogBlockRules installs a block-rule catalog for one test and
// restores the previous hook afterwards, the way keywords_test.go
// stubs CatalogPrintedKeywords.
func stubCatalogBlockRules(t *testing.T, fn func(key string) []BlockRule) {
	t.Helper()
	prev := CatalogBlockRules
	t.Cleanup(func() { CatalogBlockRules = prev })
	CatalogBlockRules = fn
}

// carryBlockRule stamps blockRuleOracle onto a battlefield permanent,
// making it the source the stubbed hook answers for.
func carryBlockRule(t *testing.T, g *Game, card uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, card)
		if c == nil {
			t.Fatalf("setup: %s is not on the battlefield", card)
		}
		c.OracleID = blockRuleOracle
	})
}

// cantBeBlockedByWeaklings is Legolas Greenleaf's shape: a pair rule on
// the attacker itself ("this creature can't be blocked by creatures
// with power 2 or less"). The source check is what makes it a rule
// about ITS OWN attacker rather than about every attacker on the
// board.
func cantBeBlockedByWeaklings() []BlockRule {
	return []BlockRule{{
		Reason: BlockReasonCantBeBlockedBy,
		Label:  "creatures with power 2 or less",
		Pair: func(g *Game, attacker, blocker, source *Card) bool {
			if attacker == nil || blocker == nil || source == nil {
				return false
			}
			if attacker.InstanceID != source.InstanceID {
				return false
			}
			return blocker.PowerForComparison() <= 2
		},
	}}
}

// TestBlockRuleRefusesThePairAtDeclaration — the headline of #750. A
// conditional restriction is enforced where every other pair rule is,
// so DeclareBlocker refuses the block instead of accepting one the
// rules forbid, and the refusal carries a sentence naming the clause.
func TestBlockRuleRefusesThePairAtDeclaration(t *testing.T) {
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key != blockRuleOracle {
			return nil
		}
		return cantBeBlockedByWeaklings()
	})

	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Legolas Greenleaf", 3, 3)
	weak := pushCombatant(t, g, g.Seats[1], "Weakling", 2, 2)
	strong := pushCombatant(t, g, g.Seats[1], "Bruiser", 3, 3)
	carryBlockRule(t, g, attacker)
	declareAttacks(t, g, attacker)

	err := g.DeclareBlocker(weak, attacker)
	if err == nil {
		t.Fatalf("a power-2 blocker was accepted against a creature that can't be blocked by one")
	}
	if !errors.Is(err, ErrIllegalBlock) {
		t.Errorf("refusal does not wrap ErrIllegalBlock: %v", err)
	}
	var refused *BlockRefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("refusal is not a *BlockRefusedError: %T", err)
	}
	if refused.Reason != BlockReasonCantBeBlockedBy {
		t.Errorf("reason %q, want %q", refused.Reason, BlockReasonCantBeBlockedBy)
	}
	if refused.Source != attacker {
		t.Errorf("the refusal names %s as its source, want the rule's permanent %s", refused.Source, attacker)
	}
	if refused.Label != "creatures with power 2 or less" {
		t.Errorf("the refusal dropped the rule's printed clause: %q", refused.Label)
	}
	if s := refused.Sentence(uuid.Nil); s == "" {
		t.Error("the refusal has no player-facing sentence")
	}
	if c := findCard(g, weak); c == nil || c.BlockingTarget != uuid.Nil {
		t.Error("the refused blocker was staged anyway")
	}

	// The same rule lets a big enough creature through, which is what
	// makes it conditional rather than "can't be blocked".
	if err := g.DeclareBlocker(strong, attacker); err != nil {
		t.Fatalf("a power-3 blocker was refused: %v", err)
	}
}

// TestBlockRuleOnlyBindsItsOwnAttacker — the `source` argument is the
// scope. A rule read from one permanent says nothing about a second
// attacker standing next to it.
func TestBlockRuleOnlyBindsItsOwnAttacker(t *testing.T) {
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key != blockRuleOracle {
			return nil
		}
		return cantBeBlockedByWeaklings()
	})

	g := newActiveGame(t)
	ruled := pushCombatant(t, g, g.Seats[0], "Legolas Greenleaf", 3, 3)
	plain := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	weak := pushCombatant(t, g, g.Seats[1], "Weakling", 2, 2)
	carryBlockRule(t, g, ruled)
	declareAttacks(t, g, ruled, plain)

	if err := g.DeclareBlocker(weak, plain); err != nil {
		t.Fatalf("the rule refused a block against an attacker it says nothing about: %v", err)
	}
}

// TestBlockRuleStopsWhenItsSourceLeaves — the rule is a static ability
// of a permanent (CR 113.6), so it is read from the battlefield every
// time a block is checked rather than copied onto the attacker. A
// Champion-of-Lambholt-shaped rule dies with its lord.
func TestBlockRuleStopsWhenItsSourceLeaves(t *testing.T) {
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key != blockRuleOracle {
			return nil
		}
		return []BlockRule{{
			Reason: BlockReasonCantBlockAttacker,
			Label:  "creatures with power less than this creature's power",
			Pair: func(g *Game, attacker, blocker, source *Card) bool {
				if attacker == nil || blocker == nil || source == nil {
					return false
				}
				// Only creatures the lord's controller controls are
				// protected, which is the printed clause.
				if attacker.Controller != source.Controller {
					return false
				}
				return blocker.PowerForComparison() < source.PowerForComparison()
			},
		}}
	})

	g := newActiveGame(t)
	lord := pushCombatant(t, g, g.Seats[0], "Champion of Lambholt", 4, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	blocker := pushCombatant(t, g, g.Seats[1], "Small Blocker", 1, 4)
	carryBlockRule(t, g, lord)
	declareAttacks(t, g, attacker)

	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Fatalf("the lord's rule did not refuse a smaller blocker")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(lord); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("the rule outlived the permanent that printed it: %v", err)
	}
}

// TestBlockRuleMaximumRefusesAnOverfullBlockAtDeclaration — Hungering
// Hydra's shape. A maximum is a bound on the whole declaration, so it
// is judged where the declaration is complete: DeclareBlockers, which
// REFUSES the over-full set whole rather than storing part of it
// (#750, Decision 13). Refusing the set whole is deliberate — which
// of the two blockers to drop is the defender's choice, not the
// engine's.
func TestBlockRuleMaximumRefusesAnOverfullBlockAtDeclaration(t *testing.T) {
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key != blockRuleOracle {
			return nil
		}
		return []BlockRule{{
			Label: "more than one creature",
			Count: func(g *Game, attacker, source *Card) (int, int) {
				if attacker == nil || source == nil || attacker.InstanceID != source.InstanceID {
					return 0, 0
				}
				return 0, 1
			},
		}}
	})

	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Hungering Hydra", 4, 4)
	first := pushCombatant(t, g, g.Seats[1], "Blocker One", 1, 4)
	second := pushCombatant(t, g, g.Seats[1], "Blocker Two", 1, 4)
	carryBlockRule(t, g, attacker)
	declareAttacks(t, g, attacker)

	err := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: first, Attacker: attacker},
		{Blocker: second, Attacker: attacker},
	})
	var refusal *BlockRefusedError
	if !errors.As(err, &refusal) {
		t.Fatalf("a two-creature block against a maximum of one = %v, want a refusal", err)
	}
	if refusal.Reason != BlockReasonTooManyBlockers || refusal.N != 1 {
		t.Errorf("refusal = %q with N = %d, want %q with 1", refusal.Reason, refusal.N, BlockReasonTooManyBlockers)
	}
	if got := refusal.Sentence(g.Seats[1].ID); got != "Hungering Hydra can't be blocked by more than one creature." {
		t.Errorf("the defender reads %q", got)
	}
	g.WithWriteLock(func() { g.commitBlockDeclarationLocked() })

	// All or nothing: neither blocker was stored, so neither
	// announced and the attacker never became blocked.
	for _, b := range []uuid.UUID{first, second} {
		if c := findCard(g, b); c == nil || c.BlockingTarget != uuid.Nil {
			t.Errorf("a refused block was stored anyway")
		}
		if n := len(blockDeclEvents(g, EventBlock, b)); n != 0 {
			t.Errorf("the refused block announced itself: %d block events", n)
		}
	}
	if g.blockedAttackers[attacker] {
		t.Error("an attacker whose block was refused was recorded as blocked")
	}
}

// A maximum bites the SECOND blocker too, not just a set that arrives
// over-full at once: the defender declares one legal blocker, then
// tries to gang up. The second declaration is judged against the set
// it would leave behind — one stored block plus the new one — so it
// is refused and the first block is untouched.
func TestBlockRuleMaximumRefusesASecondBlockerAddedLater(t *testing.T) {
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key != blockRuleOracle {
			return nil
		}
		return []BlockRule{{
			Count: func(g *Game, attacker, source *Card) (int, int) { return 0, 1 },
		}}
	})

	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Hungering Hydra", 4, 4)
	first := pushCombatant(t, g, g.Seats[1], "Blocker One", 1, 4)
	second := pushCombatant(t, g, g.Seats[1], "Blocker Two", 1, 4)
	carryBlockRule(t, g, attacker)
	declareAttacks(t, g, attacker)

	if err := g.DeclareBlocker(first, attacker); err != nil {
		t.Fatalf("the first blocker is within the maximum: %v", err)
	}
	err := g.DeclareBlocker(second, attacker)
	if !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("the second blocker = %v, want an illegal-block refusal", err)
	}
	if c := findCard(g, first); c == nil || c.BlockingTarget != attacker {
		t.Error("the refused second declaration disturbed the first block")
	}
	if c := findCard(g, second); c == nil || c.BlockingTarget != uuid.Nil {
		t.Error("the refused second blocker was stored anyway")
	}
}

// The other half of the same rule: one blocker is within the maximum,
// so it stands and announces normally.
func TestBlockRuleMaximumAcceptsASingleBlocker(t *testing.T) {
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key != blockRuleOracle {
			return nil
		}
		return []BlockRule{{
			Count: func(g *Game, attacker, source *Card) (int, int) { return 0, 1 },
		}}
	})

	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Hungering Hydra", 4, 4)
	lone := pushCombatant(t, g, g.Seats[1], "Blocker One", 1, 4)
	carryBlockRule(t, g, attacker)
	declareAttacks(t, g, attacker)

	if err := g.DeclareBlocker(lone, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	g.WithWriteLock(func() { g.commitBlockDeclarationLocked() })

	if c := findCard(g, lone); c == nil || c.BlockingTarget != attacker {
		t.Fatal("a legal single block was reverted")
	}
	if n := len(blockDeclEvents(g, EventBlock, lone)); n != 1 {
		t.Errorf("the legal block announced %d times, want 1", n)
	}
	if !g.blockedAttackers[attacker] {
		t.Error("the attacker was not recorded as blocked")
	}
}

// TestMenaceLoneBlockFiresNoBlockTriggers is the defect #750 describes,
// pinned from the trigger side. An illegal lone block against a menace
// attacker is now REFUSED at declaration, so it is never stored and
// never reaches the lock-in — "whenever this creature blocks"
// (CR 509.3a) and "becomes blocked" (CR 506.4) cannot fire on it.
// Under CR 733.1 an illegal declaration is rewound and nothing
// triggers from it; refusing it outright is the same answer arrived
// at before the defender was ever told the block was good.
func TestMenaceLoneBlockFiresNoBlockTriggers(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Menacer", 3, 3, "menace")
	lone := pushCombatant(t, g, g.Seats[1], "Lone Blocker", 2, 2)
	declareAttacks(t, g, attacker)

	if err := g.DeclareBlocker(lone, attacker); !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("DeclareBlocker on a menace attacker = %v, want an illegal-block refusal", err)
	}
	g.WithWriteLock(func() { g.commitBlockDeclarationLocked() })

	if c := findCard(g, lone); c == nil || c.BlockingTarget != uuid.Nil {
		t.Fatalf("the illegal lone block against menace was stored (CR 509.1b)")
	}
	if n := len(blockDeclEvents(g, EventBlock, lone)); n != 0 {
		t.Errorf("a reverted block fired %d 'whenever this blocks' events (CR 509.3a), want 0", n)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attacker)); n != 0 {
		t.Errorf("a reverted block made the attacker 'become blocked' %d times, want 0", n)
	}
	if g.blockedAttackers[attacker] {
		t.Error("a menace attacker with one blocker was recorded as blocked")
	}
}

// A legal menace block is untouched by the same close-out: both
// blockers announce, and the attacker becomes blocked once (CR 506.4).
func TestMenaceTwoBlockersAnnounceNormally(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Menacer", 3, 3, "menace")
	first := pushCombatant(t, g, g.Seats[1], "Blocker One", 1, 4)
	second := pushCombatant(t, g, g.Seats[1], "Blocker Two", 1, 4)
	declareAttacks(t, g, attacker)

	if err := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: first, Attacker: attacker},
		{Blocker: second, Attacker: attacker},
	}); err != nil {
		t.Fatalf("a legal two-creature menace block was refused: %v", err)
	}
	for _, b := range []uuid.UUID{first, second} {
		if c := findCard(g, b); c == nil || c.BlockingTarget != attacker {
			t.Fatalf("a legal menace block was not stored")
		}
	}
	g.WithWriteLock(func() { g.commitBlockDeclarationLocked() })

	for _, b := range []uuid.UUID{first, second} {
		if n := len(blockDeclEvents(g, EventBlock, b)); n != 1 {
			t.Errorf("each blocker of a legal menace block blocks once: %d events", n)
		}
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attacker)); n != 1 {
		t.Errorf("a menace block is one 'becomes blocked': %d events", n)
	}
}

// TestBlockerBoundsCombineMenaceAndARuleMaximum — ADR 0045 addendum,
// Decision 12: the effective minimum is the largest minimum and the
// effective maximum is the smallest maximum, and a minimum above the
// maximum makes the attacker unblockable (a menace creature wearing
// Vorrac Battlehorns). "No blockers" stays legal, as it always is.
func TestBlockerBoundsCombineMenaceAndARuleMaximum(t *testing.T) {
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key != blockRuleOracle {
			return nil
		}
		return []BlockRule{{
			Count: func(g *Game, attacker, source *Card) (int, int) { return 0, 1 },
		}}
	})

	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Menacer", 3, 3, "menace")
	carryBlockRule(t, g, attacker)
	declareAttacks(t, g, attacker)

	var lowest, highest int
	var none, one, two bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		atk := findBattlefieldCard(g, attacker)
		lowest, highest = g.blockerBoundsLocked(atk)
		none = g.blockerCountValidLocked(atk, 0)
		one = g.blockerCountValidLocked(atk, 1)
		two = g.blockerCountValidLocked(atk, 2)
	})

	if lowest != 2 {
		t.Errorf("menace minimum is %d, want 2 (CR 702.111b)", lowest)
	}
	if highest != 1 {
		t.Errorf("rule maximum is %d, want 1", highest)
	}
	if !none {
		t.Error("declining to block is always legal")
	}
	if one || two {
		t.Error("a minimum above the maximum leaves no legal block at all")
	}
}
