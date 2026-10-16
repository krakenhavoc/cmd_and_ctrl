package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shriveling_rot_test.go — #3997: a repeating delayed trigger that
// triggers once for each creature dealt damage at the same time (CR
// 603.2c, 120.4b, 510.2), and one for each creature that dies.

const shrivelingRotOracle = "9ce8691a-d9e2-45b2-bda0-2c3891b55307"

// srCast casts Shriveling Rot from the active player's hand with the
// given bullets, entwined or not, and resolves it.
func srCast(t *testing.T, g *game.Game, me *game.Player, modes []int, entwined bool) uuid.UUID {
	t.Helper()
	pool := "{C}{C}{B}{B}"
	if entwined {
		pool = "{C}{C}{C}{C}{B}{B}{B}"
	}
	id, err := castEntwineSpell(g, me, "Shriveling Rot", "Instant", "{2}{B}{B}", shrivelingRotOracle,
		[]string{"B"}, pool, modes, entwined, nil)
	if err != nil {
		t.Fatalf("cast Shriveling Rot: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// srTriggers counts the instances of Shriveling Rot's delayed triggers
// whose label contains `what`, pending or on the stack, and lists the
// creatures they are about.
func srTriggers(g *game.Game, spell uuid.UUID, what string) []uuid.UUID {
	var out []uuid.UUID
	add := func(item *game.StackItem) {
		if item == nil || item.SourceCardID != spell || !strings.Contains(item.Label, what) {
			return
		}
		for _, ref := range item.Payload {
			if ref.Kind == game.TargetCard {
				out = append(out, ref.ID)
			}
		}
	}
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered {
			add(item)
		}
	}
	for _, item := range g.PendingTriggers {
		add(item)
	}
	return out
}

func srCount(ids []uuid.UUID, id uuid.UUID) int {
	n := 0
	for _, x := range ids {
		if x == id {
			n++
		}
	}
	return n
}

func TestShrivelingRotIsRegisteredFullWithEntwine(t *testing.T) {
	spec, ok := Lookup(shrivelingRotOracle)
	if !ok || spec.Name != "Shriveling Rot" {
		t.Fatal("Shriveling Rot is not registered")
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("Completeness %v, want Full", spec.Completeness)
	}
	if offers := game.OptionalCostsFor(shrivelingRotOracle); len(offers) != 1 || offers[0].Key != game.EntwineKey {
		t.Errorf("optional costs %+v, want one entwine cost", offers)
	}
}

// Two sources dealing damage to one creature in one damage instruction
// are one occurrence: one trigger for it. A second creature in the same
// instruction is its own occurrence.
func TestShrivelingRotTriggersOnceForEachCreatureDealtDamageTogether(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	spell := srCast(t, g, me, []int{0}, false)
	big := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 9, 9)
	other := b12Creature(g, opp.ID, "Other Bear", "Creature — Bear", 9, 9)
	srcA := b12Creature(g, me.ID, "Source A", "Creature — Bear", 2, 2)
	srcB := b12Creature(g, me.ID, "Source B", "Creature — Bear", 3, 3)

	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(srcA, big, 2)
			_ = g.DealDamageToCreatureForEffect(srcB, big, 3)
			return g.DealDamageToCreatureForEffect(srcA, other, 1)
		})
	})
	g.RunStateChecksForTest()
	got := srTriggers(g, spell, "destroy")
	if srCount(got, big) != 1 || srCount(got, other) != 1 || len(got) != 2 {
		t.Fatalf("triggers about %v, want one for each damaged creature", got)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, big) || onBattlefield(g, other) {
		t.Error("both damaged creatures are destroyed")
	}
}

// Two separate damage instructions are two events, so they are two
// occurrences even inside one batch (ADR 0108 amendment 2026-10-13).
func TestShrivelingRotSeparateInstructionsAreSeparateTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	spell := srCast(t, g, me, []int{0}, false)
	big := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 9, 9)
	src := b12Creature(g, me.ID, "Source", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(src, big, 2)
		_ = g.DealDamageToCreatureForEffect(src, big, 3)
	})
	g.RunStateChecksForTest()
	if got := srTriggers(g, spell, "destroy"); srCount(got, big) != 2 {
		t.Fatalf("triggers about %v, want two for the two instructions", got)
	}
}

// A Fiery Confluence that chooses "1 damage to each creature" three times
// deals damage three separate times (CR 608.2c), so a creature triggers it
// three times: two regeneration shields are not enough.
func TestShrivelingRotRepeatedModeIsThreeOccurrences(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	srCast(t, g, me, []int{0}, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 9)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for range 2 {
			if err := g.RegenerateForEffect(bear); err != nil {
				t.Fatalf("RegenerateForEffect: %v", err)
			}
		}
	})
	castCatalogSpellWithModes(t, g, "Fiery Confluence", "Sorcery", "3c22e031-4804-4c31-bd3c-c3f29d456b34", []int{0, 0, 0})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, bear) {
		t.Error("three destroys use up two shields and destroy the bear")
	}
}

// The headline case: an attacker blocked by two creatures is dealt damage
// by both in one combat damage step, so it is destroyed once, and one
// regeneration shield saves it. The blocker it dealt damage is its own
// occurrence.
func TestShrivelingRotDoubleBlockedAttackerTriggersOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	spell := srCast(t, g, me, []int{0}, false)
	ogre := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 6)
	bearA := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 1, 4)
	bearB := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 1, 4)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if err := g.RegenerateForEffect(ogre); err != nil {
			t.Fatalf("RegenerateForEffect: %v", err)
		}
	})

	declareAttack(t, g, opp.ID, ogre)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bearA, ogre); err != nil {
		t.Fatalf("DeclareBlocker A: %v", err)
	}
	if err := g.DeclareBlocker(bearB, ogre); err != nil {
		t.Fatalf("DeclareBlocker B: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	b26AnswerDamageAssignments(t, g)
	g.RunStateChecksForTest()

	got := srTriggers(g, spell, "destroy")
	if srCount(got, ogre) != 1 {
		t.Fatalf("triggers about %v: the Ogre is one occurrence, want 1 for it", got)
	}
	// The Ogre's 4 damage all goes to Bear A (b26AnswerDamageAssignments),
	// which is its own occurrence; Bear B was dealt none.
	if srCount(got, bearA) != 1 || srCount(got, bearB) != 0 || len(got) != 2 {
		t.Errorf("triggers about %v: want one for the Ogre and one for Bear A", got)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, ogre) {
		t.Error("the Ogre's one regeneration shield saves it from the one destroy")
	}
}

// "Destroy it" is that object (CR 603.7c): a creature that left and came
// back before the trigger resolved is a new object and stays.
func TestShrivelingRotDoesNotDestroyTheCreatureAsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	spell := srCast(t, g, me, []int{0}, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 9)
	src := b12Creature(g, me.ID, "Source", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, bear, 1) })
	if got := srTriggers(g, spell, "destroy"); srCount(got, bear) != 1 {
		t.Fatalf("triggers about %v, want one for the bear", got)
	}
	var returned uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bear); err != nil {
			t.Fatalf("exile: %v", err)
		}
		back, err := g.ReturnFromExileToBattlefieldForEffect(bear, opp.ID, false)
		if err != nil || back == uuid.Nil {
			t.Fatalf("return: %v", err)
		}
		returned = back
	})
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, returned) {
		t.Error("the flickered bear is a new object and is not destroyed")
	}
}

// The second bullet: a creature that dies costs its controller life equal
// to its toughness as it last existed, counters included (CR 608.2h).
func TestShrivelingRotDiesCostsTheControllerItsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	spell := srCast(t, g, me, []int{1}, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if _, err := g.AddCounterMustSettleNowForEffect(bear, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if got := srTriggers(g, spell, "toughness"); srCount(got, bear) != 1 {
		t.Fatalf("triggers about %v, want one for the bear", got)
	}
	life, mine := opp.Life, me.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 || me.Life != mine {
		t.Errorf("the bear's controller lost %d (want 3), the caster lost %d (want 0)", life-opp.Life, mine-me.Life)
	}
}

// Entwined: the creature dealt damage is destroyed, and its death costs
// its controller its toughness. Both triggers end with the turn.
func TestShrivelingRotEntwinedDestroysThenDrains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	srCast(t, g, me, []int{0, 1}, true)
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("%d delayed triggers, want 2", len(g.DelayedTriggers))
	}
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 5)
	src := b12Creature(g, me.ID, "Source", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, bear, 1) })
	life := opp.Life
	passPriorityAroundTable(t, g)
	if onBattlefield(g, bear) {
		t.Fatal("the bear dealt damage is destroyed")
	}
	if opp.Life != life-5 {
		t.Errorf("the bear's controller lost %d, want 5", life-opp.Life)
	}
	advanceToUpkeepOf(t, g, 1)
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers survived the turn, want 0", len(g.DelayedTriggers))
	}
}
