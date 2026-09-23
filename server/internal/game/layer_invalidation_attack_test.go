package game

import (
	"testing"

	"github.com/google/uuid"
)

// layer_invalidation_attack_test.go covers #1218: the ATTACKING
// STATUS quarter of "Layer cache invalidation on hand / life / attack
// / graveyard state" that #1117 left open (see docs/engine-seams.md).
// Hand size and life total were fixed by #74 and #1117; the graveyard
// half by #1117 too. Nothing bumped when a creature was declared as
// an attacker, was removed from combat mid-declaration, or when
// combat ended — so a static reading "is this creature attacking"
// (Ohran Frostfang's "attacking creatures you control have
// deathtouch") could survive stale across all three.
//
// As with the life-total tests, every assertion here is made with NO
// other event in between: a battlefield move would drop the cached
// resolution on its own and hide a missing bump.

const attackStatusStaticOracle = "attack-status-static"

// attackStatusGatedGrantForTest is a stub in the Ohran Frostfang
// shape: a Layer 6 keyword grant to attacking creatures the source's
// controller controls, declaring the dependency so the listener
// knows to invalidate.
func attackStatusGatedGrantForTest(declare bool) StaticAbility {
	return StaticAbility{
		Layer:                    Layer6Ability,
		DependsOnAttackingStatus: declare,
		AppliesTo: func(target *Card, _ *Game, source *Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && target.AttackingTarget != uuid.Nil
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			c.Abilities = append(c.Abilities, "deathtouch")
		},
	}
}

// pushAttackStatusGatedCreature seeds a creature carrying the stub
// and returns it.
func pushAttackStatusGatedCreature(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	return pushTypedTestCard(g, Card{
		Name:       "Frostfang Stand-In",
		TypeLine:   "Creature — Snake",
		OracleID:   attackStatusStaticOracle,
		Power:      2,
		Toughness:  6,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
}

// --- the declare-attackers gate --------------------------------------

// TestLayerVersionBumpsOnEventAttackWhenAnAttackingStaticIsLive is the
// declare-attackers third of the fix.
func TestLayerVersionBumpsOnEventAttackWhenAnAttackingStaticIsLive(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == attackStatusStaticOracle {
			return []StaticAbility{attackStatusGatedGrantForTest(true)}
		}
		return nil
	})
	g := newActiveGame(t)
	pushAttackStatusGatedCreature(t, g, g.Seats[0])
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	before := readLayerVersion(g)
	lockInAttackDeclaration(t, g)
	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump on the attack declaration's lock-in while an attacking-keyed static was in play: was %d, now %d", before, got)
	}
}

// TestLayerVersionIgnoresEventAttackForAStaticThatDoesNotDeclareIt is
// the guard: a board with no attacking-keyed static pays nothing for
// a declare-attackers step.
func TestLayerVersionIgnoresEventAttackForAStaticThatDoesNotDeclareIt(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == attackStatusStaticOracle {
			return []StaticAbility{attackStatusGatedGrantForTest(false)}
		}
		return nil
	})
	g := newActiveGame(t)
	pushAttackStatusGatedCreature(t, g, g.Seats[0])
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	before := readLayerVersion(g)
	lockInAttackDeclaration(t, g)
	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion bumped on an attack declaration with no attacking-keyed static in play: was %d, now %d", before, got)
	}
}

// --- the removed-from-combat gate ------------------------------------
//
// A control change (materialiseControlLocked, layers.go) is the
// printed way a permanent leaves combat mid-declaration (CR 506.4),
// but EventControlChanged already carries its own UNCONDITIONAL bump
// (#990) — so measuring layerVersion around one cannot isolate this
// gate's contribution. Regeneration's "remove it from combat" (CR
// 701.19a, #921) goes through the very same removeFromCombatLocked
// call with no event of its own (EventRegenerated is not on the
// listener's switch at all), which is what makes it the clean vehicle
// here.

// TestLayerVersionBumpsWhenRegenerationRemovesAnAttackerFromCombat is
// the CR 506.4 exit via CR 701.19a: regeneration clears
// Card.AttackingTarget with no event of its own.
func TestLayerVersionBumpsWhenRegenerationRemovesAnAttackerFromCombat(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == attackStatusStaticOracle {
			return []StaticAbility{attackStatusGatedGrantForTest(true)}
		}
		return nil
	})
	g := newActiveGame(t)
	pushAttackStatusGatedCreature(t, g, g.Seats[0])
	me, opp := g.Seats[0], g.Seats[1]
	id := regenBear(g, me.ID)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		c.AttackingTarget = opp.ID
	})
	regenerate(t, g, id)

	before := readLayerVersion(g)
	destroy(t, g, id)
	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump when regeneration removed an attacker from combat while an attacking-keyed static was in play: was %d, now %d", before, got)
	}
}

// TestLayerVersionIgnoresCombatRemovalForAStaticThatDoesNotDeclareIt
// is the guard for the same route.
func TestLayerVersionIgnoresCombatRemovalForAStaticThatDoesNotDeclareIt(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == attackStatusStaticOracle {
			return []StaticAbility{attackStatusGatedGrantForTest(false)}
		}
		return nil
	})
	g := newActiveGame(t)
	pushAttackStatusGatedCreature(t, g, g.Seats[0])
	me, opp := g.Seats[0], g.Seats[1]
	id := regenBear(g, me.ID)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		c.AttackingTarget = opp.ID
	})
	regenerate(t, g, id)

	before := readLayerVersion(g)
	destroy(t, g, id)
	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion bumped on a regeneration-driven combat removal with no attacking-keyed static in play: was %d, now %d", before, got)
	}
}

// --- the end-of-combat gate -------------------------------------------

// TestLayerVersionBumpsWhenCombatEndsWithAnAttackerInPlay is the CR
// 511.3 exit: clearCombatLocked wipes every AttackingTarget on the
// board in one sweep, with no event of its own.
func TestLayerVersionBumpsWhenCombatEndsWithAnAttackerInPlay(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == attackStatusStaticOracle {
			return []StaticAbility{attackStatusGatedGrantForTest(true)}
		}
		return nil
	})
	g := newActiveGame(t)
	pushAttackStatusGatedCreature(t, g, g.Seats[0])
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)

	before := readLayerVersion(g)
	if err := g.ClearCombat(); err != nil {
		t.Fatalf("ClearCombat: %v", err)
	}
	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump when combat ended with an attacker in play while an attacking-keyed static was in play: was %d, now %d", before, got)
	}
}

// TestLayerVersionIgnoresEndOfCombatForAStaticThatDoesNotDeclareIt is
// the guard for the same route, and also covers the no-attackers
// no-op: PassTurn calls clearCombatLocked every turn whether or not
// combat happened at all.
func TestLayerVersionIgnoresEndOfCombatForAStaticThatDoesNotDeclareIt(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == attackStatusStaticOracle {
			return []StaticAbility{attackStatusGatedGrantForTest(false)}
		}
		return nil
	})
	g := newActiveGame(t)
	pushAttackStatusGatedCreature(t, g, g.Seats[0])
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)

	before := readLayerVersion(g)
	if err := g.ClearCombat(); err != nil {
		t.Fatalf("ClearCombat: %v", err)
	}
	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion bumped on end of combat with no attacking-keyed static in play: was %d, now %d", before, got)
	}
}

// TestLayerVersionIgnoresEndOfCombatWithNoAttackers is the same guard
// from the other end: a combat (or a turn with no combat at all) that
// never had an attacker invalidates nothing, even with the static
// live, because there is nothing for the sweep to change.
func TestLayerVersionIgnoresEndOfCombatWithNoAttackers(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == attackStatusStaticOracle {
			return []StaticAbility{attackStatusGatedGrantForTest(true)}
		}
		return nil
	})
	g := newActiveGame(t)
	pushAttackStatusGatedCreature(t, g, g.Seats[0])

	before := readLayerVersion(g)
	if err := g.ClearCombat(); err != nil {
		t.Fatalf("ClearCombat: %v", err)
	}
	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion bumped on end of combat with no attacker ever declared: was %d, now %d", before, got)
	}
}
