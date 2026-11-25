package game

import (
	"testing"

	"github.com/google/uuid"
)

// block_reach_waiver_test.go — #6250: "This creature can block Dragons
// as though it had reach" (Dragon Hunter). CR 702.9b and CR 702.17b say
// a creature with flying can't be blocked except by creatures with
// flying and/or reach; the waiver bends that for the attackers the
// clause names and for nothing else.

// dragonQuery is "Dragons", the clause Dragon Hunter prints.
func dragonQuery() PermanentQuery {
	return PermanentQuery{Types: []string{"creature"}, Subtypes: []string{"Dragon"}}
}

// reachWaiverBoard is a flying Dragon and a flying Angel attacking a
// defender whose ground Dragon Hunter carries the stubbed rule.
func reachWaiverBoard(t *testing.T) (g *Game, dragon, angel, hunter uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	dragon = pushTypedTestCard(g, Card{
		Name: "Test Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4,
		Owner: me, Controller: me, Keywords: []string{"flying"},
	})
	angel = pushTypedTestCard(g, Card{
		Name: "Test Angel", TypeLine: "Creature — Angel", Power: 4, Toughness: 4,
		Owner: me, Controller: me, Keywords: []string{"flying"},
	})
	hunter = pushTypedTestCard(g, Card{
		Name: "Dragon Hunter", TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 1,
		Owner: them, Controller: them,
	})
	// A ground attacker too, so the defender always owes a block
	// decision and the step does not finish their declaration for them
	// before the waiver is stubbed.
	ground := pushTypedTestCard(g, Card{
		Name: "Ground Attacker", TypeLine: "Creature — Test", Power: 1, Toughness: 1,
		Owner: me, Controller: me,
	})
	for _, id := range []uuid.UUID{dragon, angel, hunter, ground} {
		if c := findCard(g, id); c != nil {
			c.SummonedThisTurn = false
		}
	}
	attackAndStepToBlocks(t, g, them, dragon, angel, ground)
	return g, dragon, angel, hunter
}

// TestReachWaiverBlocksTheNamedAttackerOnly — CR 702.17b: the Hunter
// blocks a flying Dragon and not a flying Angel.
func TestReachWaiverBlocksTheNamedAttackerOnly(t *testing.T) {
	g, dragon, angel, hunter := reachWaiverBoard(t)
	if r := pairRefusal(g, dragon, hunter); r.Reason != BlockReasonFlying {
		t.Fatalf("before the waiver: reason = %q, want flying", r.Reason)
	}
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key == blockRuleOracle {
			return []BlockRule{{ReachAgainst: []PermanentQuery{dragonQuery()}}}
		}
		return nil
	})
	carryBlockRule(t, g, hunter)
	if r := pairRefusal(g, dragon, hunter); !r.Legal() {
		t.Fatalf("blocking the flying Dragon: refused with %q, want legal", r.Reason)
	}
	if err := g.DeclareBlocker(hunter, dragon); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if r := pairRefusal(g, angel, hunter); r.Reason != BlockReasonFlying {
		t.Fatalf("blocking the flying Angel: reason = %q, want flying", r.Reason)
	}
}

// TestReachWaiverIsNotReach — the creature does not gain the keyword:
// the waiver is read at the block check, and nothing that asks about
// reach sees it.
func TestReachWaiverIsNotReach(t *testing.T) {
	g, _, _, hunter := reachWaiverBoard(t)
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key == blockRuleOracle {
			return []BlockRule{{ReachAgainst: []PermanentQuery{dragonQuery()}}}
		}
		return nil
	})
	carryBlockRule(t, g, hunter)
	if HasKeyword(findCard(g, hunter), "reach") {
		t.Error("the waiver granted the reach keyword")
	}
}

// TestReachWaiverIsOnlyTheBlockersOwnRule — a waiver on another
// permanent grants nothing: the clause is a creature's own ability about
// its own blocks.
func TestReachWaiverIsOnlyTheBlockersOwnRule(t *testing.T) {
	g, dragon, _, hunter := reachWaiverBoard(t)
	them := g.Seats[1].ID
	bystander := pushTypedTestCard(g, Card{
		Name: "Bystander", TypeLine: "Enchantment", Owner: them, Controller: them,
	})
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key == blockRuleOracle {
			return []BlockRule{{ReachAgainst: []PermanentQuery{dragonQuery()}}}
		}
		return nil
	})
	carryBlockRule(t, g, bystander)
	if r := pairRefusal(g, dragon, hunter); r.Reason != BlockReasonFlying {
		t.Fatalf("a waiver on another permanent: reason = %q, want flying", r.Reason)
	}
}

// TestReachWaiverLeavesOtherEvasion — the waiver bends CR 702.9b and
// nothing else: a flying Dragon with shadow is still unblockable
// (CR 509.1b, different evasion abilities are cumulative).
func TestReachWaiverLeavesOtherEvasion(t *testing.T) {
	g, dragon, _, hunter := reachWaiverBoard(t)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, dragon)
		eff := c.Effective()
		eff.Abilities = append(eff.Abilities, "shadow")
		c.effective = &eff
	})
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key == blockRuleOracle {
			return []BlockRule{{ReachAgainst: []PermanentQuery{dragonQuery()}}}
		}
		return nil
	})
	carryBlockRule(t, g, hunter)
	if r := pairRefusal(g, dragon, hunter); r.Reason != BlockReasonShadow {
		t.Fatalf("a shadow Dragon: reason = %q, want shadow", r.Reason)
	}
}
