package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ohran_frostfang_test.go — #1218: the card proof for the
// attacking-status quarter of layer invalidation
// (docs/engine-seams.md, "Layer cache invalidation on hand, life,
// attack and graveyard events"). Every assertion here is made with NO
// other event between the state change and the read, which is the
// property that was missing: a stale cache survives just fine until
// something unrelated happens to invalidate it, and that is exactly
// what hid this gap.

const ohranFrostfangOracle = "b99ada26-9a61-4175-9fb8-15a106960220"

// TestOhranFrostfangGrantsDeathtouchOnlyToAttackers is the
// declare-attackers half: the static must read the FRESH attacking
// status the instant the declaration locks in, not whatever the cache
// held before it.
func TestOhranFrostfangGrantsDeathtouchOnlyToAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ohran Frostfang", "Snow Creature — Snake", ohranFrostfangOracle, false)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	homebody := pushVanillaCreature(g, me.ID, "Homebody", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)

	if got := effectiveAbilities(t, g, attacker); !containsString(got, "deathtouch") {
		t.Errorf("attacker's abilities = %v, want deathtouch", got)
	}
	if got := effectiveAbilities(t, g, homebody); containsString(got, "deathtouch") {
		t.Errorf("a creature that stayed home has deathtouch: %v", got)
	}
}

// TestOhranFrostfangLosesDeathtouchWhenCombatEnds is the CR 511.3
// exit: clearCombatLocked wipes every AttackingTarget on the board in
// one sweep with no event of its own, and the cache must not survive
// it.
func TestOhranFrostfangLosesDeathtouchWhenCombatEnds(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ohran Frostfang", "Snow Creature — Snake", ohranFrostfangOracle, false)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if got := effectiveAbilities(t, g, attacker); !containsString(got, "deathtouch") {
		t.Fatalf("attacker never had deathtouch to lose: %v", got)
	}

	if err := g.ClearCombat(); err != nil {
		t.Fatalf("ClearCombat: %v", err)
	}
	if got := effectiveAbilities(t, g, attacker); containsString(got, "deathtouch") {
		t.Errorf("deathtouch survived the end of combat: %v", got)
	}
}

// TestOhranFrostfangDrawsOnCombatDamageToAPlayer is the mandatory
// draw half — no "you may" the way Bident of Thassa's is.
func TestOhranFrostfangDrawsOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ohran Frostfang", "Snow Creature — Snake", ohranFrostfangOracle, false)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)

	before := me.Hand.Size()
	attackWith(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size = %d, want %d (one card drawn off the combat-damage trigger)", got, before+1)
	}
}
