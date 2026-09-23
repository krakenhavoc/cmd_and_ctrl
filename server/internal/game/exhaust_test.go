package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// exhaust_test.go — #1181, the "Activate each exhaust ability only
// once" half of the exhaust keyword.
//
// What is pinned here is the engine half: the record is per ABILITY
// and per OBJECT, it is written at the ANNOUNCE, it never refreshes
// for the object that spent it, it survives an undo and a snapshot,
// and a new object (a flicker, a copy) gets its own. The enumerator's
// and the wire's halves are in internal/legal and internal/protocol,
// both reading the same Game.AbilityExhausted.

const (
	exhaustLabelA = "Exhaust — {1}: probe A"
	exhaustLabelB = "Exhaust — {1}: probe B"
	exhaustLabelC = "Exhaust — {1}: probe C"
	plainLabel    = "{1}: plain, repeatable"
)

// exhaustProbe is one ability shape whose effect stamps a counter, so
// a test can tell an activation that went through from one that was
// refused.
func exhaustProbe(label string, exhaust bool) ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label:   label,
		Exhaust: exhaust,
		Cost:    AbilityCost{Mana: "{1}"},
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "ran-"+label, 1)
		},
	}
}

// exhaustCatalog installs a catalog entry for `oracle` for the length
// of the test.
//
// A catalog hook rather than Card.ActivatedAbilities, and the reason
// is the snapshot: an intrinsic ability list is closures, which no
// restore can bring back, so a card that carried its abilities on the
// instance would come out of a round trip with no abilities at all and
// the round-trip test would pass for the wrong reason. Every printed
// exhaust card is a catalog card anyway.
func exhaustCatalog(t *testing.T, oracle string, abilities ...ActivatedAbilityShape) {
	t.Helper()
	prev := CatalogActivatedAbilities
	CatalogActivatedAbilities = func(key string) []ActivatedAbilityShape {
		if key == oracle {
			return abilities
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogActivatedAbilities = prev })
}

// pushExhaustSource puts an artifact keyed on `oracle` on the
// battlefield.
func pushExhaustSource(g *Game, owner *Player, oracle string) uuid.UUID {
	c := NewCard("Exhaust Probe", owner.ID)
	c.TypeLine = "Artifact"
	c.OracleID = oracle
	c.Controller = owner.ID
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// fireExhaust activates ability `index` of `src` with one generic
// mana floated for it.
func fireExhaust(t *testing.T, g *Game, me *Player, src uuid.UUID, index int) error {
	t.Helper()
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	err := g.ActivateCatalogAbility(me.ID, src, index, ActivateAbilityParams{})
	if err != nil {
		// A refusal pays nothing, so the mana is still floating and
		// would otherwise leak into the next activation.
		me.ManaPool = nil
	}
	return err
}

func exhaustedThisGame(g *Game, src uuid.UUID, label string) int {
	var n int
	g.WithWriteLock(func() { n = g.ActivatedThisGame(src, label) })
	return n
}

// TestAnExhaustAbilityCanBeActivatedOnlyOnce is the headline, and the
// second half is the one that matters: the refusal pays nothing.
func TestAnExhaustAbilityCanBeActivatedOnlyOnce(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-one", exhaustProbe(exhaustLabelA, true))
	src := pushExhaustSource(g, me, "probe-one")

	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	passBothForTest(g)
	if got := counterOf(g, src, "ran-"+exhaustLabelA); got != 1 {
		t.Fatalf("the first activation resolved %d times, want 1", got)
	}

	stackBefore := len(g.StackMeta)
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{})
	if !errors.Is(err, ErrAbilityExhausted) {
		t.Fatalf("second activation: err = %v, want ErrAbilityExhausted", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %v, want the {C} still floating — a refused activation pays nothing", me.ManaPool)
	}
	if len(g.StackMeta) != stackBefore {
		t.Error("the refused activation put something on the stack")
	}
}

// TestThreeExhaustAbilitiesOnOneCardAreIndependent — the record is per
// ABILITY. Loot, the Pathfinder prints three of them and may activate
// each once; spending the first must not lock the other two.
func TestThreeExhaustAbilitiesOnOneCardAreIndependent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-three",
		exhaustProbe(exhaustLabelA, true),
		exhaustProbe(exhaustLabelB, true),
		exhaustProbe(exhaustLabelC, true),
	)
	src := pushExhaustSource(g, me, "probe-three")

	for i, label := range []string{exhaustLabelA, exhaustLabelB, exhaustLabelC} {
		if err := fireExhaust(t, g, me, src, i); err != nil {
			t.Fatalf("ability %d (%s): %v", i, label, err)
		}
		passBothForTest(g)
	}
	for i, label := range []string{exhaustLabelA, exhaustLabelB, exhaustLabelC} {
		if got := counterOf(g, src, "ran-"+label); got != 1 {
			t.Errorf("ability %d resolved %d times, want 1", i, got)
		}
		if err := fireExhaust(t, g, me, src, i); !errors.Is(err, ErrAbilityExhausted) {
			t.Errorf("ability %d re-activation: err = %v, want ErrAbilityExhausted", i, err)
		}
	}
}

// TestAnExhaustAbilityDoesNotLockThePermanentsOtherAbilities is the
// same property from the other side, on the shape Greenbelt Guardian
// prints: one exhaust ability beside one ordinary repeatable one.
func TestAnExhaustAbilityDoesNotLockThePermanentsOtherAbilities(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-mixed",
		exhaustProbe(plainLabel, false),
		exhaustProbe(exhaustLabelA, true),
	)
	src := pushExhaustSource(g, me, "probe-mixed")

	if err := fireExhaust(t, g, me, src, 1); err != nil {
		t.Fatalf("the exhaust ability: %v", err)
	}
	passBothForTest(g)
	for i := 0; i < 3; i++ {
		if err := fireExhaust(t, g, me, src, 0); err != nil {
			t.Fatalf("plain activation %d after the exhaust was spent: %v", i, err)
		}
		passBothForTest(g)
	}
	if got := counterOf(g, src, "ran-"+plainLabel); got != 3 {
		t.Errorf("the repeatable ability resolved %d times, want 3 — spending an exhaust "+
			"ability says nothing about the permanent's other abilities", got)
	}
	if err := fireExhaust(t, g, me, src, 1); !errors.Is(err, ErrAbilityExhausted) {
		t.Errorf("the exhaust ability is still spent: err = %v, want ErrAbilityExhausted", err)
	}
}

// TestAnExhaustAbilityIsSpentAtTheANNOUNCE — an exhaust ability
// countered on the stack, or one that never resolves for any other
// reason, has still been activated (CR 602.2b: activating is one
// indivisible step and it finished).
func TestAnExhaustAbilityIsSpentAtTheANNOUNCE(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-announce", exhaustProbe(exhaustLabelA, true))
	src := pushExhaustSource(g, me, "probe-announce")

	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := exhaustedThisGame(g, src, exhaustLabelA); got != 1 {
		t.Fatalf("the record says %d activations with the ability still on the stack, want 1", got)
	}
	// Counter it, which is the printed reason the record cannot live
	// at resolution.
	var itemID uuid.UUID
	for id := range g.StackMeta {
		itemID = id
	}
	if itemID == uuid.Nil {
		t.Fatal("setup: the activation put nothing on the stack")
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(itemID); err != nil {
			t.Fatalf("counter the ability: %v", err)
		}
	})
	passBothForTest(g)
	if got := counterOf(g, src, "ran-"+exhaustLabelA); got != 0 {
		t.Fatalf("setup: the ability resolved after all (%d)", got)
	}
	if err := fireExhaust(t, g, me, src, 0); !errors.Is(err, ErrAbilityExhausted) {
		t.Errorf("err = %v, want ErrAbilityExhausted — a countered exhaust ability is still spent", err)
	}
}

// TestExhaustNeverRefreshesOnATurnBoundary — the game-lifetime scope.
// The per-turn half of the same record IS flushed, which is what the
// "Activate only once each turn" seam row will read.
func TestExhaustNeverRefreshesOnATurnBoundary(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-turn", exhaustProbe(exhaustLabelA, true))
	src := pushExhaustSource(g, me, "probe-turn")

	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passBothForTest(g)
	var perTurn int
	g.WithWriteLock(func() { perTurn = g.ActivatedThisTurn(src, exhaustLabelA) })
	if perTurn != 1 {
		t.Fatalf("the per-turn half reads %d, want 1", perTurn)
	}

	advanceUntil(t, g, 40, func() bool { return g.Turn.ActiveSeat == 1 && g.Turn.Step == StepPrecombatMain })

	g.WithWriteLock(func() { perTurn = g.ActivatedThisTurn(src, exhaustLabelA) })
	if perTurn != 0 {
		t.Errorf("the per-turn half reads %d on the next turn, want 0", perTurn)
	}
	if got := exhaustedThisGame(g, src, exhaustLabelA); got != 1 {
		t.Errorf("the game-lifetime half reads %d on the next turn, want 1 — exhaust never "+
			"refreshes", got)
	}
	if err := fireExhaust(t, g, me, src, 0); !errors.Is(err, ErrAbilityExhausted) {
		t.Errorf("err = %v on a later turn, want ErrAbilityExhausted", err)
	}
}

// TestAFlickerRefreshesAnExhaustAbilityAndStayingPutDoesNot is
// CR 400.7 on both sides. Leaving and returning is a new object with a
// fresh exhaust; a permanent that never changed zones keeps its spent
// one, however much else happens to it.
//
// The second half is also phasing's answer: phasing is not a zone
// change (CR 702.26d) and so does not bump Card.ObjectEpoch, so a
// phased-out permanent comes back spent. It is still written as "the
// epoch did not move" rather than as a phase-out, because the epoch is
// what the rule is about and #1199's own
// TestPhasingDoesNotBumpTheObjectEpoch asserts the phase cycle
// directly.
func TestAFlickerRefreshesAnExhaustAbilityAndStayingPutDoesNot(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-flicker", exhaustProbe(exhaustLabelA, true))
	src := pushExhaustSource(g, me, "probe-flicker")

	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passBothForTest(g)

	// Staying put: tapped, untapped, counters piled on — none of it is
	// a zone change, so none of it is a new object.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == src {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		_ = g.applyCounterLocked(src, CounterPlusOne, 1)
	})
	if err := fireExhaust(t, g, me, src, 0); !errors.Is(err, ErrAbilityExhausted) {
		t.Errorf("err = %v for a permanent that never left, want ErrAbilityExhausted", err)
	}

	epochBefore := objectEpochOf(g, src)
	bounceAndReturn(t, g, src, me.ID)
	if objectEpochOf(g, src) == epochBefore {
		t.Fatal("setup: leaving and returning did not make a new object")
	}
	if got := exhaustedThisGame(g, src, exhaustLabelA); got != 0 {
		t.Errorf("the returning object reads %d activations, want 0 — it is a new object "+
			"(CR 400.7) and its exhaust abilities are available again", got)
	}
	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Errorf("the returning object cannot use its exhaust ability: %v", err)
	}
}

// TestACopyOfThePermanentHasItsOwnExhausts — CR 707.2. What a
// permanent has already done is not a copiable value, so a copy comes
// with every exhaust ability available even when the original spent
// them.
func TestACopyOfThePermanentHasItsOwnExhausts(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-copy", exhaustProbe(exhaustLabelA, true))
	original := pushExhaustSource(g, me, "probe-copy")

	if err := fireExhaust(t, g, me, original, 0); err != nil {
		t.Fatalf("activate the original: %v", err)
	}
	passBothForTest(g)

	// The copy is what a copy is: the same printed abilities on a new
	// instance, which is a new key.
	copyID := pushExhaustSource(g, me, "probe-copy")
	if copyID == original {
		t.Fatal("setup: the copy shares the original's instance ID")
	}
	if got := exhaustedThisGame(g, copyID, exhaustLabelA); got != 0 {
		t.Fatalf("the copy reads %d activations, want 0", got)
	}
	if err := fireExhaust(t, g, me, copyID, 0); err != nil {
		t.Errorf("the copy cannot use its own exhaust ability: %v", err)
	}
	if err := fireExhaust(t, g, me, original, 0); !errors.Is(err, ErrAbilityExhausted) {
		t.Errorf("the original: err = %v, want ErrAbilityExhausted — the copy spent its own", err)
	}
}

// TestExhaustSurvivesUndo — a rewind puts the record back where it
// was. An undo that kept the activation would take an ability away for
// the rest of the game on the strength of something that no longer
// happened.
func TestExhaustSurvivesUndo(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-undo", exhaustProbe(exhaustLabelA, true))
	src := pushExhaustSource(g, me, "probe-undo")

	before := g.Clone()
	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := exhaustedThisGame(g, src, exhaustLabelA); got != 1 {
		t.Fatalf("before the undo: %d, want 1", got)
	}

	g.RestoreFrom(before)
	if got := exhaustedThisGame(g, src, exhaustLabelA); got != 0 {
		t.Errorf("after undoing the activation: %d, want 0", got)
	}
	// And the clone is not sharing the live game's map.
	if got := exhaustedThisGame(before, src, exhaustLabelA); got != 0 {
		t.Errorf("the snapshot itself reads %d, want 0 — the maps must be copied, not shared", got)
	}
}

// TestExhaustSurvivesASnapshotRoundTrip — "have I used this yet" is
// game state a player can lose a game over, so it is `carried`
// (#1020's plan row) and a restored game refuses the second activation
// exactly as the live one does.
func TestExhaustSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-snapshot", exhaustProbe(exhaustLabelA, true))
	src := pushExhaustSource(g, me, "probe-snapshot")
	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passBothForTest(g)

	_, restored := roundTrip(t, g)
	if got := exhaustedThisGame(restored, src, exhaustLabelA); got != 1 {
		t.Fatalf("the restored game reads %d activations, want 1", got)
	}
	restoredMe := restored.Seats[0]
	if err := fireExhaust(t, restored, restoredMe, src, 0); !errors.Is(err, ErrAbilityExhausted) {
		t.Errorf("the restored game: err = %v, want ErrAbilityExhausted", err)
	}
}

// objectEpochOf reads a card's CR 400.7 object serial number.
func objectEpochOf(g *Game, id uuid.UUID) int {
	var n int
	g.WithWriteLock(func() { n = g.objectEpochLocked(id) })
	return n
}
