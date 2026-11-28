package game

import (
	"testing"

	"github.com/google/uuid"
)

// block_shadow_waiver_test.go — S82 OLD batch 767 (#5628): "This creature can
// block creatures with shadow as though it had shadow" (Wall of Diffusion).
// CR 702.28b says a creature with shadow can't be blocked by creatures without
// shadow; the waiver (BlockRule.ShadowAgainstShadow) bends that and nothing
// else. The twin of block_reach_waiver_test.go.

func stubShadowWaiver(t *testing.T) {
	t.Helper()
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key == blockRuleOracle {
			return []BlockRule{{ShadowAgainstShadow: true}}
		}
		return nil
	})
}

// shadowWaiverBoard is a shadow attacker, a shadow-and-flying attacker and a
// plain attacker against a defender's ground Wall that will carry the stubbed
// rule.
func shadowWaiverBoard(t *testing.T) (g *Game, shade, flyingShade, plain, wall uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	shade = pushTypedTestCard(g, Card{
		Name: "Test Shade", TypeLine: "Creature — Shade", Power: 2, Toughness: 2,
		Owner: me, Controller: me, Keywords: []string{"shadow"},
	})
	flyingShade = pushTypedTestCard(g, Card{
		Name: "Test Flying Shade", TypeLine: "Creature — Shade", Power: 2, Toughness: 2,
		Owner: me, Controller: me, Keywords: []string{"shadow", "flying"},
	})
	plain = pushTypedTestCard(g, Card{
		Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me, Controller: me,
	})
	wall = pushTypedTestCard(g, Card{
		Name: "Wall of Diffusion", TypeLine: "Creature — Wall", Power: 0, Toughness: 5,
		Owner: them, Controller: them,
	})
	for _, id := range []uuid.UUID{shade, flyingShade, plain, wall} {
		if c := findCard(g, id); c != nil {
			c.SummonedThisTurn = false
		}
	}
	attackAndStepToBlocks(t, g, them, shade, flyingShade, plain)
	return g, shade, flyingShade, plain, wall
}

// TestShadowWaiverBlocksAShadowAttacker — CR 702.28b: with the waiver the Wall
// blocks the shadow attacker, and without it the pair is refused for shadow.
func TestShadowWaiverBlocksAShadowAttacker(t *testing.T) {
	g, shade, _, plain, wall := shadowWaiverBoard(t)
	if r := pairRefusal(g, shade, wall); r.Reason != BlockReasonShadow {
		t.Fatalf("before the waiver: reason = %q, want shadow", r.Reason)
	}
	stubShadowWaiver(t)
	carryBlockRule(t, g, wall)
	if r := pairRefusal(g, shade, wall); !r.Legal() {
		t.Fatalf("blocking the shadow attacker: refused with %q, want legal", r.Reason)
	}
	if r := pairRefusal(g, plain, wall); !r.Legal() {
		t.Fatalf("blocking an attacker without shadow: refused with %q, want legal as always", r.Reason)
	}
	if err := g.DeclareBlocker(wall, shade); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
}

// TestShadowWaiverIsNotShadow — the creature does not gain the keyword.
func TestShadowWaiverIsNotShadow(t *testing.T) {
	g, _, _, _, wall := shadowWaiverBoard(t)
	stubShadowWaiver(t)
	carryBlockRule(t, g, wall)
	if HasKeyword(findCard(g, wall), "shadow") {
		t.Error("the waiver granted the shadow keyword")
	}
}

// TestShadowWaiverIsOnlyTheBlockersOwnRule — a waiver on another permanent
// grants nothing.
func TestShadowWaiverIsOnlyTheBlockersOwnRule(t *testing.T) {
	g, shade, _, _, wall := shadowWaiverBoard(t)
	them := g.Seats[1].ID
	bystander := pushTypedTestCard(g, Card{
		Name: "Bystander", TypeLine: "Enchantment", Owner: them, Controller: them,
	})
	stubShadowWaiver(t)
	carryBlockRule(t, g, bystander)
	if r := pairRefusal(g, shade, wall); r.Reason != BlockReasonShadow {
		t.Fatalf("a waiver on another permanent: reason = %q, want shadow", r.Reason)
	}
}

// TestShadowWaiverLeavesOtherEvasion — the waiver bends CR 702.28b and nothing
// else: a flying shadow attacker still can't be blocked by a ground Wall.
func TestShadowWaiverLeavesOtherEvasion(t *testing.T) {
	g, _, flyingShade, _, wall := shadowWaiverBoard(t)
	stubShadowWaiver(t)
	carryBlockRule(t, g, wall)
	if r := pairRefusal(g, flyingShade, wall); r.Reason != BlockReasonFlying {
		t.Fatalf("a flying shadow attacker: reason = %q, want flying", r.Reason)
	}
}
