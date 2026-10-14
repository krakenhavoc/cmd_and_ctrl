package game

import (
	"testing"

	"github.com/google/uuid"
)

// assigns_no_combat_damage_test.go pins ADR 0187 §4 (#3569, CR 510.1a):
// "this creature assigns no combat damage this turn". It is a rule about
// assignment, not a prevention shield, so every test here is a statement
// about what the combat damage pass computes for a marked combatant.

// noDamageCreature is a creature whose combat keywords are baked into its
// effective characteristic. It carries no battlefield-entry stamp, so it is
// not summoning sick and a pinned record names it at epoch 0.
func noDamageCreature(t *testing.T, g *Game, owner *Player, power, toughness int, keywords ...string) uuid.UUID {
	t.Helper()
	return pushKeywordCreature(t, g, owner, power, toughness, keywords...)
}

// markNoCombatDamage is "<id> assigns no combat damage this turn".
func markNoCombatDamage(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() { ok = g.AssignsNoCombatDamageThisTurnForEffect(uuid.Nil, id, "Carrion Rats") })
	if !ok {
		t.Fatalf("AssignsNoCombatDamageThisTurnForEffect(%s) registered nothing", id)
	}
	if !g.PermanentAssignsNoCombatDamageForEffect(id) {
		t.Fatalf("the record is not read back for %s", id)
	}
}

// attackOneBlock attacks seat 1 with `atk`, blocks with each of `blockers`
// and steps into the combat damage step.
func attackOneBlock(t *testing.T, g *Game, atk uuid.UUID, blockers ...uuid.UUID) {
	t.Helper()
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	for _, b := range blockers {
		if err := g.DeclareBlocker(b, atk); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepCombatDamage)
}

func damageMarkedOn(t *testing.T, g *Game, id uuid.UUID) int {
	t.Helper()
	c := findCard(g, id)
	if c == nil {
		t.Fatalf("card %s left the battlefield", id)
	}
	return c.DamageMarked
}

// The headline: unblocked, the marked attacker deals nothing to the
// defending player, and the same creature without the mark does.
func TestAssignsNoCombatDamageUnblockedDealsNothing(t *testing.T) {
	for _, tc := range []struct {
		name     string
		mark     bool
		wantLife int
	}{
		{"marked", true, 40},
		{"control, no mark", false, 37},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			rats := noDamageCreature(t, g, g.Seats[0], 3, 3)
			if tc.mark {
				markNoCombatDamage(t, g, rats)
			}
			attackOneBlock(t, g, rats)
			if g.Seats[1].Life != tc.wantLife {
				t.Errorf("defender life = %d, want %d", g.Seats[1].Life, tc.wantLife)
			}
		})
	}
}

// Blocked by one creature, the marked attacker marks no damage on it, and
// the blocker still hits back.
func TestAssignsNoCombatDamageBlockedMarksNothing(t *testing.T) {
	g := newActiveGame(t)
	rats := noDamageCreature(t, g, g.Seats[0], 3, 3)
	ogre := noDamageCreature(t, g, g.Seats[1], 2, 5)
	markNoCombatDamage(t, g, rats)
	attackOneBlock(t, g, rats, ogre)
	if n := damageMarkedOn(t, g, ogre); n != 0 {
		t.Errorf("blocker has %d damage, want 0", n)
	}
	if n := damageMarkedOn(t, g, rats); n != 2 {
		t.Errorf("the attacker took %d from the blocker, want 2 (the mark stops only its own damage)", n)
	}
}

// Multi-blocked, there is nothing to divide (CR 510.1a): no
// damage_assignment prompt is queued. The control queues one.
func TestAssignsNoCombatDamageMultiBlockedQueuesNoPrompt(t *testing.T) {
	for _, tc := range []struct {
		name       string
		mark       bool
		wantPrompt bool
	}{
		{"marked", true, false},
		{"control, no mark", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			rats := noDamageCreature(t, g, g.Seats[0], 5, 6)
			bear1 := noDamageCreature(t, g, g.Seats[1], 2, 2)
			bear2 := noDamageCreature(t, g, g.Seats[1], 3, 3)
			if tc.mark {
				markNoCombatDamage(t, g, rats)
			}
			attackOneBlock(t, g, rats, bear1, bear2)
			var prompt bool
			for _, pc := range g.PendingChoices {
				if pc.Kind == PendingChoiceDamageAssignment {
					prompt = true
				}
			}
			if prompt != tc.wantPrompt {
				t.Errorf("damage_assignment prompt queued = %v, want %v", prompt, tc.wantPrompt)
			}
			if !tc.mark {
				return
			}
			for _, b := range []uuid.UUID{bear1, bear2} {
				if n := damageMarkedOn(t, g, b); n != 0 {
					t.Errorf("a blocker has %d damage, want 0", n)
				}
			}
			if n := damageMarkedOn(t, g, rats); n != 5 {
				t.Errorf("the attacker took %d from its blockers, want 5", n)
			}
		})
	}
}

// Trample, lifelink and deathtouch cause nothing: there is no damage for
// them to act on. Commander damage is the same.
func TestAssignsNoCombatDamageKeywordsCauseNothing(t *testing.T) {
	t.Run("blocked", func(t *testing.T) {
		g := newActiveGame(t)
		wurm := noDamageCreature(t, g, g.Seats[0], 6, 6, "trample", "lifelink", "deathtouch")
		bear := noDamageCreature(t, g, g.Seats[1], 3, 3)
		markNoCombatDamage(t, g, wurm)
		startLife := g.Seats[0].Life
		attackOneBlock(t, g, wurm, bear)
		if findCard(g, bear) == nil {
			t.Error("deathtouch destroyed the blocker")
		}
		if g.Seats[1].Life != 40 {
			t.Errorf("trample overflow: defender life = %d, want 40", g.Seats[1].Life)
		}
		if g.Seats[0].Life != startLife {
			t.Errorf("lifelink gained %d life, want 0", g.Seats[0].Life-startLife)
		}
	})
	t.Run("unblocked", func(t *testing.T) {
		g := newActiveGame(t)
		wurm := noDamageCreature(t, g, g.Seats[0], 6, 6, "trample", "lifelink", "deathtouch")
		markNoCombatDamage(t, g, wurm)
		startLife := g.Seats[0].Life
		attackOneBlock(t, g, wurm)
		if g.Seats[1].Life != 40 || g.Seats[0].Life != startLife {
			t.Errorf("life = %d / %d, want 40 / %d", g.Seats[1].Life, g.Seats[0].Life, startLife)
		}
	})
}

// "Assigns damage equal to its toughness" (Ghalta-style, CR 510.1c)
// cannot lift the 0: the mark is read ahead of combatDamagePower.
func TestAssignsNoCombatDamageBeatsDamageByToughness(t *testing.T) {
	const oracle = "assigns-no-damage-toughness-wall"
	withStaticAbilities(t, func(id string) []StaticAbility {
		if id != oracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.DamageByToughness = true
				c.Abilities = append(c.Abilities, "haste")
			},
		}}
	})
	g := newActiveGame(t)
	wall := pushTypedTestCard(g, Card{
		Name: "Toughness Wall", TypeLine: "Creature — Wall", OracleID: oracle,
		Power: 0, Toughness: 7, Owner: g.Seats[0].ID, Controller: g.Seats[0].ID,
	})
	markNoCombatDamage(t, g, wall)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	var power int
	g.ReadSnapshot(func() { power = findCard(g, wall).combatDamagePower() })
	if power != 7 {
		t.Fatalf("setup: DamageByToughness assigns %d, want 7", power)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if g.Seats[1].Life != 40 {
		t.Errorf("defender life = %d, want 40", g.Seats[1].Life)
	}
}

// A blocker with the mark deals no damage to the attacker; it is still hit.
func TestAssignsNoCombatDamageMarkedBlockerDealsNothing(t *testing.T) {
	g := newActiveGame(t)
	bear := noDamageCreature(t, g, g.Seats[0], 2, 2)
	rats := noDamageCreature(t, g, g.Seats[1], 3, 3)
	markNoCombatDamage(t, g, rats)
	attackOneBlock(t, g, bear, rats)
	if n := damageMarkedOn(t, g, bear); n != 0 {
		t.Errorf("attacker has %d damage, want 0", n)
	}
	if n := damageMarkedOn(t, g, rats); n != 2 {
		t.Errorf("marked blocker took %d, want 2", n)
	}
}

// It is not prevention (CR 615): "damage can't be prevented" does not
// undo it, because nothing is assigned for the effect to protect.
func TestAssignsNoCombatDamageIsNotPrevention(t *testing.T) {
	g := newActiveGame(t)
	rats := noDamageCreature(t, g, g.Seats[0], 3, 3)
	ogre := noDamageCreature(t, g, g.Seats[1], 1, 5)
	markNoCombatDamage(t, g, rats)
	g.WithWriteLock(func() {
		if !g.DamageCantBePreventedThisTurnForEffect(uuid.Nil, "Skullcrack") {
			t.Fatal("setup: the can't-be-prevented grant registered nothing")
		}
	})
	attackOneBlock(t, g, rats, ogre)
	if n := damageMarkedOn(t, g, ogre); n != 0 {
		t.Errorf("blocker has %d damage under an unpreventable-damage grant, want 0", n)
	}
	if n := damageMarkedOn(t, g, rats); n != 1 {
		t.Errorf("the attacker took %d, want the blocker's 1", n)
	}
}

// CR 514.2: "this turn" ends at cleanup.
func TestAssignsNoCombatDamageEndsAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	rats := noDamageCreature(t, g, g.Seats[0], 3, 3)
	markNoCombatDamage(t, g, rats)
	g.WithWriteLock(func() { g.sweepTurnEndLocked() })
	if g.PermanentAssignsNoCombatDamageForEffect(rats) {
		t.Error("the mark survived the cleanup sweep")
	}
	attackOneBlock(t, g, rats)
	if g.Seats[1].Life != 37 {
		t.Errorf("defender life = %d, want 37 once the turn has ended", g.Seats[1].Life)
	}
}

// CR 400.7: the record names the object. A creature that left and came
// back is a new object and assigns damage normally.
func TestAssignsNoCombatDamageDoesNotFollowAFlicker(t *testing.T) {
	g := newActiveGame(t)
	rats := noDamageCreature(t, g, g.Seats[0], 3, 3)
	markNoCombatDamage(t, g, rats)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rats {
				g.Battlefield.Cards[i].EnteredBattlefieldAt += 1_000_000
			}
		}
	})
	if g.PermanentAssignsNoCombatDamageForEffect(rats) {
		t.Error("the new object still carries the old object's mark")
	}
	attackOneBlock(t, g, rats)
	if g.Seats[1].Life != 37 {
		t.Errorf("defender life = %d, want 37", g.Seats[1].Life)
	}
}

// A card that is not on the battlefield gets no record.
func TestAssignsNoCombatDamageRegistersNothingOffTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	before := len(g.ScopedEffects)
	var ok bool
	g.WithWriteLock(func() { ok = g.AssignsNoCombatDamageThisTurnForEffect(uuid.Nil, uuid.New(), "Carrion Rats") })
	if ok || len(g.ScopedEffects) != before {
		t.Errorf("ok=%v, %d records (was %d); want false and no new record", ok, len(g.ScopedEffects), before)
	}
	if g.PermanentAssignsNoCombatDamageForEffect(uuid.New()) {
		t.Error("an unknown card reads as marked")
	}
}
