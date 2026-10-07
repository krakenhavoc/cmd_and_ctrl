package game

import (
	"testing"

	"github.com/google/uuid"
)

// planeswalker_creature_test.go — #2046: a planeswalker that becomes a
// creature and is still a planeswalker (the Gideons). Damage to it has
// BOTH results (CR 120.3c loyalty off, CR 120.3e damage marked) and
// BOTH state-based actions apply (CR 704.5g lethal damage, CR 704.5i no
// loyalty). ADR 0032, amendment of 2026-10-07.

// animateWalkerForTest makes `id` a power/toughness creature that is
// still a planeswalker until end of turn, the shape every Gideon's
// animate ability registers.
func animateWalkerForTest(t *testing.T, g *Game, id uuid.UUID, power, toughness int, keywords ...string) {
	t.Helper()
	mods := []Mod{AddTypesMod("Creature"), SetBasePowerMod(power), SetBaseToughnessMod(toughness)}
	if len(keywords) > 0 {
		mods = append(mods, AddKeywordsMod(keywords...))
	}
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(id, g.PinnedObjectsLocked(id), mods,
			g.UntilEndOfTurnDuration(), "test — becomes a creature that's still a planeswalker")
		g.RecomputeLayersIfStaleLocked()
	})
	c := findBattlefieldCard(g, id)
	if c == nil || !c.IsCreature() || !c.IsPlaneswalker() {
		t.Fatalf("setup: %v is not both a creature and a planeswalker", id)
	}
	if c.CurrentToughness() != toughness {
		t.Fatalf("setup: toughness = %d, want %d", c.CurrentToughness(), toughness)
	}
}

// dealUnpreventable deals `amount` damage that can't be prevented to the
// permanent from a fresh red source and runs the state-based actions.
func dealUnpreventable(t *testing.T, g *Game, target uuid.UUID, amount int) {
	t.Helper()
	bolt := pushRedSpell(g, g.Seats[0])
	g.WithWriteLock(func() {
		if err := g.DealMarkedDamageForEffect(bolt, nil, target, amount, DamageMarks{CantBePrevented: true}); err != nil {
			t.Fatalf("damage: %v", err)
		}
		g.runStateChecksLocked()
	})
}

// The damage has both results (CR 120.3c and 120.3e): loyalty comes off
// AND the damage is marked, and a hit that leaves both thresholds
// unreached leaves the permanent standing.
func TestCreatureWalkerTakesBothDamageResults(t *testing.T) {
	g := newActiveGame(t)
	gideon := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Gideon", 6)
	animateWalkerForTest(t, g, gideon, 6, 6)

	dealUnpreventable(t, g, gideon, 2)

	c := findBattlefieldCard(g, gideon)
	if c == nil {
		t.Fatal("a 6-loyalty, 6-toughness Gideon died to 2 damage")
	}
	if got := c.Counters[CounterLoyalty]; got != 4 {
		t.Errorf("loyalty = %d, want 4 (CR 120.3c)", got)
	}
	if c.DamageMarked != 2 {
		t.Errorf("damage marked = %d, want 2 (CR 120.3e)", c.DamageMarked)
	}
}

// CR 704.5i for an object that is also a creature: the damage takes the
// last loyalty counter, the creature rules are not reached (3 damage on
// toughness 4), and the permanent still goes to the graveyard. Before
// #2046 `stateBasedActionsLocked` skipped the planeswalker rule for any
// creature, so this Gideon stayed on the battlefield at 0 loyalty.
func TestCreatureWalkerAtNoLoyaltyIsPutIntoTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	gideon := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Gideon", 3)
	animateWalkerForTest(t, g, gideon, 4, 4)

	dealUnpreventable(t, g, gideon, 3)

	if g.Battlefield.Contains(gideon) {
		t.Fatal("a creature planeswalker with no loyalty counters stayed on the battlefield (CR 704.5i)")
	}
	if !g.Seats[1].Graveyard.Contains(gideon) {
		t.Error("Gideon is not in his owner's graveyard")
	}
}

// CR 704.5g for an object that is also a planeswalker: lethal damage
// destroys it however many loyalty counters it has.
func TestCreatureWalkerWithLethalDamageIsDestroyed(t *testing.T) {
	g := newActiveGame(t)
	gideon := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Gideon", 9)
	animateWalkerForTest(t, g, gideon, 4, 4)

	dealUnpreventable(t, g, gideon, 4)

	if g.Battlefield.Contains(gideon) {
		t.Fatal("a creature planeswalker with lethal damage stayed on the battlefield (CR 704.5g)")
	}
	if !g.Seats[1].Graveyard.Contains(gideon) {
		t.Error("Gideon is not in his owner's graveyard")
	}
}

// Indestructible stops 704.5g and nothing else. Lethal damage marked on
// an indestructible Gideon with loyalty left is survived ...
func TestIndestructibleCreatureWalkerSurvivesLethalDamageWithLoyaltyLeft(t *testing.T) {
	g := newActiveGame(t)
	gideon := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Gideon", 9)
	animateWalkerForTest(t, g, gideon, 4, 4, "indestructible")

	dealUnpreventable(t, g, gideon, 4)

	c := findBattlefieldCard(g, gideon)
	if c == nil {
		t.Fatal("an indestructible Gideon with 5 loyalty left was put into the graveyard")
	}
	if c.DamageMarked != 4 || c.Counters[CounterLoyalty] != 5 {
		t.Errorf("damage = %d, loyalty = %d; want 4 and 5", c.DamageMarked, c.Counters[CounterLoyalty])
	}
}

// ... but "put into a graveyard" is not destruction (CR 704.5i), so the
// same indestructible Gideon loses his last loyalty counter and goes.
func TestIndestructibleCreatureWalkerStillDiesAtNoLoyalty(t *testing.T) {
	g := newActiveGame(t)
	gideon := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Gideon", 4)
	animateWalkerForTest(t, g, gideon, 4, 4, "indestructible")

	dealUnpreventable(t, g, gideon, 4)

	if g.Battlefield.Contains(gideon) {
		t.Fatal("indestructible kept a planeswalker with no loyalty counters on the battlefield (CR 704.5i is not destruction)")
	}
}

// Both rules apply to one check (CR 704.3): the permanent leaves ONCE,
// in one move, not once per rule.
func TestCreatureWalkerDoomedByBothRulesLeavesOnce(t *testing.T) {
	g := newActiveGame(t)
	gideon := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Gideon", 4)
	animateWalkerForTest(t, g, gideon, 4, 4)

	dealUnpreventable(t, g, gideon, 4) // loyalty 0 AND lethal damage

	n := 0
	for _, c := range g.Seats[1].Graveyard.Cards {
		if c.InstanceID == gideon {
			n++
		}
	}
	if n != 1 || g.Battlefield.Contains(gideon) {
		t.Errorf("Gideon appears %d times in the graveyard (on battlefield: %v), want exactly 1 and gone",
			n, g.Battlefield.Contains(gideon))
	}
}

// A creature whose toughness the engine doesn't know skips the creature
// rules whole (Card.ToughnessIsKnown) -- but its loyalty is still read.
func TestCreatureWalkerWithNoLoyaltyDiesEvenWhenToughnessIsUnknown(t *testing.T) {
	g := newActiveGame(t)
	c := NewCard("Starred Walker", g.Seats[1].ID)
	c.TypeLine = "Legendary Creature Planeswalker — Test"
	g.Battlefield.PushTop(c)
	if got := findBattlefieldCard(g, c.InstanceID); got == nil || got.ToughnessIsKnown() {
		t.Fatal("setup: the fixture's toughness should be unknown")
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if g.Battlefield.Contains(c.InstanceID) {
		t.Fatal("a creature planeswalker with unknown toughness and no loyalty stayed on the battlefield")
	}
}

// Combat damage to an attacked, animated Gideon (CR 506.4 keeps him a
// legal attack target while he is a planeswalker, CR 510.1b deals it to
// him) has both results too.
func TestCombatDamageToAnAnimatedWalkerRemovesLoyaltyAndMarksDamage(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	gideon := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Gideon", 5)
	animateWalkerForTest(t, g, gideon, 4, 4)
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, gideon); err != nil {
		t.Fatalf("DeclareAttacker at an animated walker: %v", err)
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	passUntilStep(t, g, StepEndCombat)

	c := findBattlefieldCard(g, gideon)
	if c == nil {
		t.Fatal("Gideon died to 3 combat damage at 5 loyalty and 4 toughness")
	}
	if got := c.Counters[CounterLoyalty]; got != 2 {
		t.Errorf("loyalty = %d, want 2 (CR 120.3c)", got)
	}
	if c.DamageMarked != 3 {
		t.Errorf("damage marked = %d, want 3 (CR 120.3e)", c.DamageMarked)
	}
}
