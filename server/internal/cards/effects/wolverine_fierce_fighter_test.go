package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wolverine_fierce_fighter_test.go pins Wolverine, Fierce Fighter and the
// heal-on-damage seam under it (#4614, CR 614.1a, CR 701.69a):
//
//	"If damage would be dealt to Wolverine, instead that damage is dealt,
//	 but all other damage already dealt to him is healed."

const wolverineOracleID = "8429ba6a-6c06-4a83-a6ee-023b64b825f8"

const wolverineTypeLine = "Legendary Creature — Mutant Berserker Hero"

// wolverineOnBattlefield puts a 3/5 Wolverine onto the battlefield (no
// trigger) and returns it with a quiet source that can deal damage to it.
func wolverineOnBattlefield(g *game.Game, owner, foe uuid.UUID) (wolverine, source uuid.UUID) {
	wolverine = b12Push(g, owner, "Wolverine, Fierce Fighter", wolverineTypeLine, wolverineOracleID, 3, 5)
	source = apaPush(g, foe, foe, game.Card{Name: "Hitter", TypeLine: "Creature — Test", Power: 0, Toughness: 9})
	return wolverine, source
}

func dealDamageTo(t *testing.T, g *game.Game, source, target uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(source, target, n); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
}

func TestWolverineHealsEarlierDamageSoTwoHitsOfThreeLeaveThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolv, src := wolverineOnBattlefield(g, me.ID, opp.ID)

	dealDamageTo(t, g, src, wolv, 3)
	if got := damageMarkedOn(g, wolv); got != 3 {
		t.Fatalf("after the first hit %d damage is marked, want 3", got)
	}
	dealDamageTo(t, g, src, wolv, 3)
	if !g.Battlefield.Contains(wolv) {
		t.Fatal("Wolverine died to 3 + 3: the first 3 should have been healed")
	}
	if got := damageMarkedOn(g, wolv); got != 3 {
		t.Errorf("after the second hit %d damage is marked, want 3 (the first was healed)", got)
	}
	dealDamageTo(t, g, src, wolv, 4)
	if got := damageMarkedOn(g, wolv); got != 4 || !g.Battlefield.Contains(wolv) {
		t.Errorf("after a third hit: marked %d, on battlefield %v; want 4 and alive", got, g.Battlefield.Contains(wolv))
	}
}

func TestWolverineStillDiesToOneHitOfItsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolv, src := wolverineOnBattlefield(g, me.ID, opp.ID)

	dealDamageTo(t, g, src, wolv, 3)
	dealDamageTo(t, g, src, wolv, 6)
	if g.Battlefield.Contains(wolv) {
		t.Fatal("Wolverine survived 6 damage: lethal is judged on the new damage")
	}
}

func TestWolverineLetsTheDamageThroughNoMatterWhatWasMarked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolv, src := wolverineOnBattlefield(g, me.ID, opp.ID)

	// Nothing marked: the replacement has nothing to heal and the damage is
	// dealt as it always was.
	dealDamageTo(t, g, src, wolv, 5)
	if g.Battlefield.Contains(wolv) {
		t.Fatal("5 damage to a 3/5 with nothing marked should kill it")
	}
}

func TestWolverineDeathtouchStillKills(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolv, _ := wolverineOnBattlefield(g, me.ID, opp.ID)
	snake := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Snake", TypeLine: "Creature — Snake", Power: 1, Toughness: 1,
		Keywords: []string{"deathtouch"}})

	dealDamageTo(t, g, snake, wolv, 1)
	if g.Battlefield.Contains(wolv) {
		t.Fatal("1 deathtouch damage should destroy Wolverine (CR 704.5h)")
	}
}

func TestWolverineDeathtouchAfterEarlierDamageStillKills(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolv, src := wolverineOnBattlefield(g, me.ID, opp.ID)
	snake := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Snake", TypeLine: "Creature — Snake", Power: 1, Toughness: 1,
		Keywords: []string{"deathtouch"}})

	dealDamageTo(t, g, src, wolv, 4)
	dealDamageTo(t, g, snake, wolv, 1)
	if g.Battlefield.Contains(wolv) {
		t.Fatal("deathtouch damage after healing the earlier damage should still destroy Wolverine")
	}
}

func TestWolverineWitherCountersAreNotHealedDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolv, src := wolverineOnBattlefield(g, me.ID, opp.ID)
	wither := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Withered", TypeLine: "Creature — Test", Power: 2, Toughness: 2,
		Keywords: []string{"wither"}})

	dealDamageTo(t, g, wither, wolv, 2)
	if got := findBattlefieldCardForTest(g, wolv).Counters[game.CounterMinusOne]; got != 2 {
		t.Fatalf("wither damage left %d -1/-1 counters, want 2", got)
	}
	dealDamageTo(t, g, src, wolv, 2)
	if got := findBattlefieldCardForTest(g, wolv).Counters[game.CounterMinusOne]; got != 2 {
		t.Errorf("a heal removed -1/-1 counters: %d left, want 2", got)
	}
	if got := damageMarkedOn(g, wolv); got != 2 {
		t.Errorf("%d damage marked, want 2", got)
	}
	// Wolverine is a 1/3 now, so the next 3 damage is lethal: the counters
	// stayed even though the earlier damage was healed.
	dealDamageTo(t, g, src, wolv, 3)
	if g.Battlefield.Contains(wolv) {
		t.Error("a 1/3 Wolverine should die to 3 damage")
	}
}

func TestWolverineWitherDamageHealsEarlierMarkedDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolv, src := wolverineOnBattlefield(g, me.ID, opp.ID)
	wither := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Withered", TypeLine: "Creature — Test", Power: 1, Toughness: 2,
		Keywords: []string{"wither"}})

	dealDamageTo(t, g, src, wolv, 4)
	dealDamageTo(t, g, wither, wolv, 1)
	// Wither damage is damage that would be dealt to Wolverine, so the
	// replacement applies and the 4 already dealt is healed.
	if got := damageMarkedOn(g, wolv); got != 0 {
		t.Errorf("%d damage marked after wither damage, want 0 (the earlier 4 was healed)", got)
	}
}

// One instruction's damage is dealt together (CR 510.2 for combat, one
// instruction for a spell): the first source's damage is not "already
// dealt" when the second's is.
func TestWolverineKeepsDamageDealtAtTheSameTime(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolv, src := wolverineOnBattlefield(g, me.ID, opp.ID)
	other := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Hitter Two", TypeLine: "Creature — Test", Power: 0, Toughness: 9})

	dealDamageTo(t, g, src, wolv, 2)
	g.WithWriteLock(func() {
		if err := g.DamageInstanceForEffect(func() error {
			if err := g.DealDamageToCreatureForEffect(src, wolv, 2); err != nil {
				return err
			}
			return g.DealDamageToCreatureForEffect(other, wolv, 2)
		}); err != nil {
			t.Fatal(err)
		}
	})
	// The earlier 2 was healed; the 2 + 2 dealt together stay.
	if got := damageMarkedOn(g, wolv); got != 4 {
		t.Errorf("%d damage marked, want 4 (2 + 2 dealt at once, the earlier 2 healed)", got)
	}
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(wolv) {
		t.Error("Wolverine died to 4 damage")
	}
}

func TestWolverineTwoInstructionsOfOneSpellAreTwoInstances(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolv, src := wolverineOnBattlefield(g, me.ID, opp.ID)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(src, wolv, 3); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(src, wolv, 3); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageMarkedOn(g, wolv); got != 3 {
		t.Errorf("%d damage marked, want 3: the second instruction heals the first", got)
	}
}

// combatWithTwoBlockers has Wolverine (hasty, cast this turn) attack the
// opponent and be blocked by two creatures of the given powers, then
// deals combat damage.
func combatWithTwoBlockers(t *testing.T, powerA, powerB int) (g *game.Game, wolv uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wolv = b12Push(g, me.ID, "Wolverine, Fierce Fighter", wolverineTypeLine, wolverineOracleID, 3, 5)
	b1 := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Blocker One", TypeLine: "Creature — Test", Power: powerA, Toughness: 9})
	b2 := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Blocker Two", TypeLine: "Creature — Test", Power: powerB, Toughness: 9})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wolv, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range []uuid.UUID{b1, b2} {
		if err := g.DeclareBlocker(b, wolv); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	for i := 0; i < 4; i++ {
		var c *game.PendingChoice
		for _, p := range g.PendingChoices {
			if p != nil && p.Kind == game.PendingChoiceDamageAssignment {
				c = p
			}
		}
		if c == nil {
			break
		}
		entries := []game.DamageAssignmentEntry{{BlockerID: b1, Amount: 2}, {BlockerID: b2, Amount: 1}}
		if err := g.ResolveDamageAssignment(c.ID, c.Chooser, entries, 0); err != nil {
			t.Fatalf("assign: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	return g, wolv
}

// CR 510.2: both blockers' damage is dealt at once, so neither heals the
// other: 2 + 2 is 4 marked, not 2.
func TestWolverineCombatDamageFromTwoBlockersAddsUp(t *testing.T) {
	g, wolv := combatWithTwoBlockers(t, 2, 2)
	if !g.Battlefield.Contains(wolv) {
		t.Fatal("Wolverine died to 4 combat damage")
	}
	if got := damageMarkedOn(g, wolv); got != 4 {
		t.Errorf("%d damage marked after two 2-power blockers, want 4: simultaneous damage is not healed", got)
	}
}

func TestWolverineCombatDamageFromTwoBlockersCanKillHim(t *testing.T) {
	g, wolv := combatWithTwoBlockers(t, 3, 3)
	if g.Battlefield.Contains(wolv) {
		t.Fatal("Wolverine survived 3 + 3 combat damage from two blockers at once")
	}
}

// The first-strike pass and the regular pass are two damage steps (CR 510.4),
// so the second heals the first.
func TestWolverineHealsBetweenCombatDamageSteps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wolv := b12Push(g, me.ID, "Wolverine, Fierce Fighter", wolverineTypeLine, wolverineOracleID, 3, 5)
	striker := apaPush(g, opp.ID, opp.ID, game.Card{Name: "First Striker", TypeLine: "Creature — Test", Power: 3, Toughness: 9,
		Keywords: []string{"first strike"}})
	regular := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Regular", TypeLine: "Creature — Test", Power: 3, Toughness: 9})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wolv, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range []uuid.UUID{striker, regular} {
		if err := g.DeclareBlocker(b, wolv); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	for i := 0; i < 8; i++ {
		var c *game.PendingChoice
		for _, p := range g.PendingChoices {
			if p != nil && p.Kind == game.PendingChoiceDamageAssignment {
				c = p
			}
		}
		if c == nil {
			break
		}
		entries := []game.DamageAssignmentEntry{{BlockerID: striker, Amount: 2}, {BlockerID: regular, Amount: 1}}
		if err := g.ResolveDamageAssignment(c.ID, c.Chooser, entries, 0); err != nil {
			t.Fatalf("assign: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wolv) {
		t.Fatal("Wolverine died: the first-strike 3 should have been healed by the regular 3")
	}
	if got := damageMarkedOn(g, wolv); got != 3 {
		t.Errorf("%d damage marked, want 3", got)
	}
}

func TestWolverineEntersAndFightsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 3)
	wolv := rfdCast(t, g, "Wolverine, Fierce Fighter", wolverineTypeLine, wolverineOracleID, 3, 5)

	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: wolv}); err == nil {
		t.Fatal("Wolverine may not fight himself: the trigger says \"other\"")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target survived 3 damage from the fight")
	}
	if got := damageMarkedOn(g, wolv); got != 2 {
		t.Errorf("Wolverine has %d damage marked, want 2 from the fight", got)
	}
}

// A fight is one instruction: both fighters are dealt damage together, and
// a later hit heals the fight's damage.
func TestWolverineHealsTheFightsDamageOnTheNextHit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Giant", 4, 9)
	src := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Hitter", TypeLine: "Creature — Test", Power: 0, Toughness: 9})
	wolv := rfdCast(t, g, "Wolverine, Fierce Fighter", wolverineTypeLine, wolverineOracleID, 3, 5)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, wolv); got != 4 {
		t.Fatalf("Wolverine has %d damage marked after the fight, want 4", got)
	}
	dealDamageTo(t, g, src, wolv, 2)
	if got := damageMarkedOn(g, wolv); got != 2 || !g.Battlefield.Contains(wolv) {
		t.Errorf("after a 2-damage hit: marked %d, alive %v; want 2 and alive", got, g.Battlefield.Contains(wolv))
	}
}

func TestWolverineHasHasteAndCanAttackTheTurnItEnters(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	wolv := rfdCast(t, g, "Wolverine, Fierce Fighter", wolverineTypeLine, wolverineOracleID, 3, 5)
	// No other creature is in play, so the fight has nothing to target.
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, wolv, "haste") {
		t.Fatal("Wolverine lacks haste")
	}
	life := opp.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wolv, opp.ID); err != nil {
		t.Fatalf("Wolverine could not attack the turn it entered: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3", life, opp.Life)
	}
}
