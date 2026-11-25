package game

import (
	"testing"

	"github.com/google/uuid"
)

// landwalk_ignored_test.go — #6195: "creatures with swampwalk can be
// blocked as though they didn't have swampwalk" (Quagmire, Ur-Drago)
// and "creatures with landwalk abilities can be blocked as though they
// didn't have those abilities" (Staff of the Ages). CR 702.14c makes a
// landwalker unblockable while the defending player controls a land of
// the named kind; the waiver switches that evasion off and nothing
// else (CR 509.1b: different evasion abilities are cumulative).

// waiverBoard is a swampwalker, a flying islandwalker and a plain
// creature attacking a defender who controls a Swamp and an Island,
// with one ground blocker.
func waiverBoard(t *testing.T) (g *Game, swampwalker, flyer, blocker, waiver uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	swampwalker = pushWalker(g, me, "Swampwalker", "swampwalk")
	flyer = pushWalker(g, me, "Flying Islandwalker", "flying", "islandwalk")
	blocker = pushWalker(g, them, "Blocker")
	pushLand(g, them, "Swamp", "Basic Land — Swamp")
	pushLand(g, them, "Island", "Basic Land — Island")
	waiver = pushTypedTestCard(g, Card{Name: "Waiver", TypeLine: "Enchantment", Owner: them, Controller: them})
	// A plain attacker, so the defender owes a block decision and the
	// step does not finish their declaration for them.
	plain := pushWalker(g, me, "Plain Attacker")
	attackAndStepToBlocks(t, g, them, swampwalker, flyer, plain)
	return g, swampwalker, flyer, blocker, waiver
}

// TestLandwalkWaiverMakesTheNamedWalkerBlockable — Quagmire: with the
// waiver on the battlefield a swampwalker is blockable though the
// defender controls a Swamp (CR 702.14c), and it still has the keyword.
func TestLandwalkWaiverMakesTheNamedWalkerBlockable(t *testing.T) {
	g, walker, _, blocker, waiver := waiverBoard(t)
	if r := pairRefusal(g, walker, blocker); r.Reason != BlockReasonLandwalk {
		t.Fatalf("before the waiver: reason = %q, want landwalk", r.Reason)
	}
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key == blockRuleOracle {
			return []BlockRule{{IgnoreLandwalk: []string{"swampwalk"}}}
		}
		return nil
	})
	carryBlockRule(t, g, waiver)
	if r := pairRefusal(g, walker, blocker); !r.Legal() {
		t.Fatalf("Quagmire on the table: refused with %q, want legal", r.Reason)
	}
	if err := g.DeclareBlocker(blocker, walker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if !HasKeyword(findCard(g, walker), "swampwalk") {
		t.Fatal("the waiver removed the keyword; it should only waive the evasion")
	}
}

// TestLandwalkWaiverLeavesOtherLandwalkAndFlying — the waiver names one
// landwalk: an islandwalker stays unblockable, and the CR 702.9b flying
// check is untouched (CR 702.14d: landwalk abilities don't cancel one
// another; CR 509.1b: evasion abilities are cumulative).
func TestLandwalkWaiverLeavesOtherLandwalkAndFlying(t *testing.T) {
	g, _, flyer, blocker, waiver := waiverBoard(t)
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key == blockRuleOracle {
			return []BlockRule{{IgnoreLandwalk: []string{"swampwalk"}}}
		}
		return nil
	})
	carryBlockRule(t, g, waiver)
	if r := pairRefusal(g, flyer, blocker); r.Reason != BlockReasonFlying {
		t.Fatalf("flying islandwalker: reason = %q, want flying", r.Reason)
	}
}

// TestLandwalkWaiverOfEveryKind — Staff of the Ages waives every
// landwalk ability, so an islandwalker without flying is blockable, and
// a flyer is still refused for flying.
func TestLandwalkWaiverOfEveryKind(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	islandwalker := pushWalker(g, me, "Islandwalker", "islandwalk", "snow swampwalk")
	flyer := pushWalker(g, me, "Flyer", "flying", "forestwalk")
	blocker := pushWalker(g, them, "Blocker")
	pushLand(g, them, "Island", "Basic Land — Island")
	pushLand(g, them, "Snow-Covered Swamp", "Basic Snow Land — Swamp")
	pushLand(g, them, "Forest", "Basic Land — Forest")
	staff := pushTypedTestCard(g, Card{Name: "Staff", TypeLine: "Artifact", Owner: me, Controller: me})
	attackAndStepToBlocks(t, g, them, islandwalker, flyer)
	if r := pairRefusal(g, islandwalker, blocker); r.Reason != BlockReasonLandwalk {
		t.Fatalf("before the Staff: reason = %q, want landwalk", r.Reason)
	}
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key == blockRuleOracle {
			return []BlockRule{{IgnoreAllLandwalk: true}}
		}
		return nil
	})
	carryBlockRule(t, g, staff)
	if r := pairRefusal(g, islandwalker, blocker); !r.Legal() {
		t.Fatalf("Staff of the Ages: refused with %q, want legal", r.Reason)
	}
	if r := pairRefusal(g, flyer, blocker); r.Reason != BlockReasonFlying {
		t.Fatalf("Staff of the Ages, flyer: reason = %q, want flying", r.Reason)
	}
}

// TestLandwalkWaiverIsExactToken — Quagmire waives swampwalk, not snow
// swampwalk, which is its own landwalk ability (CR 702.14c).
func TestLandwalkWaiverIsExactToken(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	walker := pushWalker(g, me, "Snow Walker", "snow swampwalk")
	blocker := pushWalker(g, them, "Blocker")
	pushLand(g, them, "Snow-Covered Swamp", "Basic Snow Land — Swamp")
	q := pushTypedTestCard(g, Card{Name: "Quagmire", TypeLine: "Enchantment", Owner: them, Controller: them})
	attackAndStepToBlocks(t, g, them, walker)
	stubCatalogBlockRules(t, func(key string) []BlockRule {
		if key == blockRuleOracle {
			return []BlockRule{{IgnoreLandwalk: []string{"swampwalk"}}}
		}
		return nil
	})
	carryBlockRule(t, g, q)
	if r := pairRefusal(g, walker, blocker); r.Reason != BlockReasonLandwalk {
		t.Fatalf("snow swampwalk under Quagmire: reason = %q, want landwalk", r.Reason)
	}
}
