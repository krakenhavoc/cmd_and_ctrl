package game

import (
	"testing"

	"github.com/google/uuid"
)

// attack_requirements_player_test.go pins the player-level CR 508.1d
// requirement (#6087): "Each opponent must attack with at least one
// creature each combat if able" (Seeker of Slaanesh) and "… attack you or
// a planeswalker you control with at least one creature …" (Trove of
// Temptation). One requirement per source, obeyed by one attacker,
// counted against every other requirement, restriction, limit and tax.

const (
	seekerTestOracle = "test-seeker-at-least-one"
	troveTestOracle  = "test-trove-at-least-one"
)

// withPlayerAttackRequirementStatics stubs the two statics the way the
// catalog writes them: onto every creature the source's opponents
// control.
func withPlayerAttackRequirementStatics(t *testing.T) {
	t.Helper()
	prev := CatalogStaticAbilities
	t.Cleanup(func() { CatalogStaticAbilities = prev })
	static := func(trove bool) StaticAbility {
		return StaticAbility{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.IsCreature() && target.Controller != source.Controller
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
				r := AttackRequirement{Source: source.InstanceID, SourceName: source.Name, AtLeastOneCreature: true}
				if trove {
					r.AttacksPlayerOrTheirPlaneswalkers = source.Controller
				}
				c.AttackRequirements = append(c.AttackRequirements, r)
			},
		}
	}
	CatalogStaticAbilities = func(key string) []StaticAbility {
		switch key {
		case seekerTestOracle:
			return []StaticAbility{static(false)}
		case troveTestOracle:
			return []StaticAbility{static(true)}
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
}

func pushSeekerForTest(g *Game, owner *Player) uuid.UUID {
	id := pushBattlefieldForTest(g, owner.ID, "Seeker", "Enchantment", seekerTestOracle)
	g.layerVersion.Add(1)
	return id
}

func pushTroveForTest(g *Game, owner *Player) uuid.UUID {
	id := pushBattlefieldForTest(g, owner.ID, "Trove", "Enchantment", troveTestOracle)
	g.layerVersion.Add(1)
	return id
}

func passOK(t *testing.T, g *Game, why string) {
	t.Helper()
	if err := g.PassPriority(); err != nil {
		t.Fatalf("%s: pass refused: %v", why, err)
	}
}

// TestPlayerAttackRequirementIsObeyedByOneAttacker — the empty
// declaration is refused while a creature could attack for free; any one
// attacker answers it, and no second creature is asked to.
func TestPlayerAttackRequirementIsObeyedByOneAttacker(t *testing.T) {
	withPlayerAttackRequirementStatics(t)
	g := newFourPlayerActiveGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	pushSeekerForTest(g, b)
	x := pushKeywordCreature(t, g, me, 2, 2)
	y := pushKeywordCreature(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	re := requirementErr(t, g.PassPriority())
	if got := re.Sentence(me.ID); got != "At least one creature must attack this combat if able (Seeker)." {
		t.Errorf("sentence = %q", got)
	}
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		must := g.MustAttackForEffect()
		if len(must[x]) == 0 || len(must[y]) == 0 {
			t.Errorf("either creature answers it; must_attack = %v", must)
		}
	})
	// The attack may go anywhere, even at a player who imposed nothing.
	if err := g.DeclareAttacker(y, c.ID); err != nil {
		t.Fatalf("declare one creature: %v", err)
	}
	passOK(t, g, "one attacker obeys it")
	if findCard(g, x).AttackingTarget != uuid.Nil {
		t.Error("the second creature was made to attack")
	}
}

// TestPlayerAttackRequirementIsNotOwedWithoutAFreeAttack — tapped,
// summoning-sick or taxed creatures owe nothing (CR 508.1d: no
// restriction is disobeyed and no cost is required).
func TestPlayerAttackRequirementIsNotOwedWithoutAFreeAttack(t *testing.T) {
	withPlayerAttackRequirementStatics(t)
	g := newFourPlayerActiveGame(t)
	me, b := g.Seats[0], g.Seats[1]
	pushSeekerForTest(g, b)
	tapped := pushKeywordCreature(t, g, me, 2, 2)
	sick := pushKeywordCreature(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.WithWriteLock(func() {
		findBattlefieldCard(g, tapped).Tapped = true
		findBattlefieldCard(g, sick).SummonedThisTurn = true
	})
	passOK(t, g, "every creature tapped or summoning-sick")

	// Taxed by every opponent: nothing is owed, and paying stays legal.
	withCatalogAttackTaxes(t, func(key string) []AttackTax {
		if key == propagandaTestOracle {
			return []AttackTax{flatTax("{2}")}
		}
		return nil
	})
	g2 := newFourPlayerActiveGame(t)
	me2 := g2.Seats[0]
	pushSeekerForTest(g2, g2.Seats[1])
	for _, p := range g2.Seats[1:] {
		pushTaxEnchantment(g2, p, "Propaganda", propagandaTestOracle)
	}
	pushKeywordCreature(t, g2, me2, 2, 2)
	advanceIntoStep(t, g2, StepDeclareAttackers)
	passOK(t, g2, "every attack taxed")
}

// TestTroveRequirementWantsItsControllerOrTheirPlaneswalker — Trove's
// shape: only an attack on its controller or a planeswalker they control
// obeys it.
func TestTroveRequirementWantsItsControllerOrTheirPlaneswalker(t *testing.T) {
	withPlayerAttackRequirementStatics(t)
	g := newFourPlayerActiveGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	pushTroveForTest(g, b)
	x := pushKeywordCreature(t, g, me, 2, 2)
	y := pushKeywordCreature(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	re := requirementErr(t, g.PassPriority())
	if got := re.Sentence(me.ID); got != "At least one creature must attack "+b.Name+" or a planeswalker they control this combat if able (Trove)." {
		t.Errorf("sentence = %q", got)
	}
	if got := re.Sentence(b.ID); got != "At least one creature must attack you or a planeswalker you control this combat if able (Trove)." {
		t.Errorf("sentence to Trove's controller = %q", got)
	}
	// An attack elsewhere is legal while another creature can still obey.
	if err := g.DeclareAttacker(x, c.ID); err != nil {
		t.Fatalf("x at another opponent: %v", err)
	}
	requirementErr(t, g.PassPriority())
	if err := g.DeclareAttacker(y, b.ID); err != nil {
		t.Fatalf("y at Trove's controller: %v", err)
	}
	passOK(t, g, "y obeys Trove")

	// One creature: it may not attack elsewhere, and Trove's controller's
	// planeswalker obeys.
	g2 := newFourPlayerActiveGame(t)
	me2, b2, c2 := g2.Seats[0], g2.Seats[1], g2.Seats[2]
	pushTroveForTest(g2, b2)
	walker := pushPlaneswalkerForTest(g2, b2.ID, "Their Walker", 4)
	otherWalker := pushPlaneswalkerForTest(g2, c2.ID, "Other Walker", 4)
	only := pushKeywordCreature(t, g2, me2, 2, 2)
	advanceIntoStep(t, g2, StepDeclareAttackers)
	requirementErr(t, g2.DeclareAttacker(only, c2.ID))
	requirementErr(t, g2.DeclareAttacker(only, otherWalker))
	if err := g2.DeclareAttacker(only, walker); err != nil {
		t.Fatalf("at Trove's controller's planeswalker: %v", err)
	}
	passOK(t, g2, "the planeswalker obeys Trove")
}

// TestPlayerAttackRequirementsAreCountedSeparately — two sources are two
// requirements, each obeyed by the same attacker; CR 508.1d makes the
// one creature go where it obeys both.
func TestPlayerAttackRequirementsAreCountedSeparately(t *testing.T) {
	withPlayerAttackRequirementStatics(t)
	g := newFourPlayerActiveGame(t)
	me, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	pushSeekerForTest(g, b)
	pushSeekerForTest(g, c)
	pushTroveForTest(g, c)
	only := pushKeywordCreature(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		r := g.attackRequirementReachLocked(me.ID, map[uuid.UUID]uuid.UUID{})
		if r.total() != 3 {
			t.Errorf("reach = %d, want 3 (two Seekers and Trove, one attacker)", r.total())
		}
	})
	requirementErr(t, g.DeclareAttacker(only, d.ID))
	requirementErr(t, g.DeclareAttacker(only, b.ID))
	if err := g.DeclareAttacker(only, c.ID); err != nil {
		t.Fatalf("at Trove's controller: %v", err)
	}
	passOK(t, g, "one attacker obeys all three")
}

// TestPlayerAttackRequirementUnderAnAttackLimit — with one slot (Silent
// Arbiter), the declaration that obeys the most requirements wins it: the
// creature that must attack, at Trove's controller.
func TestPlayerAttackRequirementUnderAnAttackLimit(t *testing.T) {
	withPlayerAttackRequirementStatics(t)
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitEachCombat, Max: 1}}, nil)
	g := newFourPlayerActiveGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	pushLimitSource(t, g, c, "Silent Arbiter")
	pushTroveForTest(g, b)
	forced := mustAttackCreature(t, g, me, 1)
	bear := pushKeywordCreature(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	requirementErr(t, g.PassPriority())
	requirementErr(t, g.DeclareAttacker(bear, b.ID))
	requirementErr(t, g.DeclareAttacker(forced, c.ID))
	if err := g.DeclareAttacker(forced, b.ID); err != nil {
		t.Fatalf("the forced creature at Trove's controller: %v", err)
	}
	passOK(t, g, "the slot obeys both")

	// Seeker alone under the limit: any one attack is the maximum, and the
	// creature with no requirement of its own may not take the slot from
	// the one that must attack.
	g2 := newFourPlayerActiveGame(t)
	me2, b2, c2 := g2.Seats[0], g2.Seats[1], g2.Seats[2]
	pushLimitSource(t, g2, c2, "Silent Arbiter")
	pushSeekerForTest(g2, b2)
	forced2 := mustAttackCreature(t, g2, me2, 1)
	bear2 := pushKeywordCreature(t, g2, me2, 2, 2)
	advanceIntoStep(t, g2, StepDeclareAttackers)
	requirementErr(t, g2.DeclareAttacker(bear2, b2.ID))
	if err := g2.DeclareAttacker(forced2, c2.ID); err != nil {
		t.Fatalf("the forced creature anywhere: %v", err)
	}
	passOK(t, g2, "one attack obeys both")
}

// TestPlayerAttackRequirementBindsOnlyOpponents — the source's
// controller is not "each opponent": their own pass is never refused.
func TestPlayerAttackRequirementBindsOnlyOpponents(t *testing.T) {
	withPlayerAttackRequirementStatics(t)
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	pushSeekerForTest(g, me)
	pushTroveForTest(g, me)
	pushKeywordCreature(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	passOK(t, g, "your own Seeker and Trove ask nothing of you")
}
