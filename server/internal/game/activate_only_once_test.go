package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// activate_only_once_test.go — #5221, CR 602.5b. The engine half of a
// plain "Activate only once": once per object, on the exhaust record,
// with none of exhaust's keyword. The cards (Possessed Goat, Haunted
// Screen, Mild-Mannered Librarian) are in
// internal/cards/effects/activate_only_once_cards_test.go.

const onlyOnceProbeLabel = "{1}: probe. Activate only once."

func onlyOnceProbe() ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label:    onlyOnceProbeLabel,
		OnlyOnce: true,
		Cost:     AbilityCost{Mana: "{1}"},
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "ran-only-once", 1)
		},
	}
}

// fireOnlyOnce floats {C} and activates ability `index` of `src`. A
// refused activation leaves the {C} floating, which the caller reads.
func fireOnlyOnce(t *testing.T, g *Game, who *Player, src uuid.UUID, index int) error {
	t.Helper()
	who.ManaPool.AddMana(ManaToken{Color: "C"})
	return g.ActivateCatalogAbility(who.ID, src, index, ActivateAbilityParams{})
}

func onlyOnceSpent(g *Game, src uuid.UUID, ab ActivatedAbilityShape) bool {
	var spent bool
	g.WithWriteLock(func() { spent = g.OnlyOnceSpent(src, ab) })
	return spent
}

func TestAnOnlyOnceAbilityIsRefusedTheSecondTime(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "only-once-second", onlyOnceProbe())
	src := pushExhaustSource(g, me, "only-once-second")

	if onlyOnceSpent(g, src, onlyOnceProbe()) {
		t.Fatal("OnlyOnceSpent before anything was activated")
	}
	if err := fireOnlyOnce(t, g, me, src, 0); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	passBothForTest(g)
	if got := counterOf(g, src, "ran-only-once"); got != 1 {
		t.Fatalf("the first activation resolved %d times, want 1", got)
	}
	err := fireOnlyOnce(t, g, me, src, 0)
	if !errors.Is(err, ErrAbilityActivatedOnce) {
		t.Fatalf("second activation: err = %v, want ErrAbilityActivatedOnce", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool = %v, want the {C} still floating — a refused activation pays nothing", me.ManaPool)
	}
	if !onlyOnceSpent(g, src, onlyOnceProbe()) {
		t.Error("OnlyOnceSpent = false after the activation")
	}
}

// TestAnOnlyOnceAbilityIsSpentAtTheAnnounce — an activation countered or
// left on the stack has still been activated, so the limit is read off
// the announcement, not the resolution.
func TestAnOnlyOnceAbilityIsSpentAtTheAnnounce(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "only-once-announce", onlyOnceProbe())
	src := pushExhaustSource(g, me, "only-once-announce")

	if err := fireOnlyOnce(t, g, me, src, 0); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if err := fireOnlyOnce(t, g, me, src, 0); !errors.Is(err, ErrAbilityActivatedOnce) {
		t.Errorf("a second activation with the first still on the stack: err = %v, want ErrAbilityActivatedOnce", err)
	}
}

// TestAnOnlyOnceAbilityIsPerAbilityAndNeverRefreshesInPlace — two such
// abilities on one card are two limits, and neither a turn boundary nor
// a change of control (CR 602.5b) gives a spent one back.
func TestAnOnlyOnceAbilityIsPerAbilityAndNeverRefreshesInPlace(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	other := onlyOnceProbe()
	other.Label = "{1}: other probe. Activate only once."
	exhaustCatalog(t, "only-once-two", onlyOnceProbe(), other)
	src := pushExhaustSource(g, me, "only-once-two")

	if err := fireOnlyOnce(t, g, me, src, 0); err != nil {
		t.Fatalf("first ability: %v", err)
	}
	passBothForTest(g)
	if onlyOnceSpent(g, src, other) {
		t.Error("activating the first ability spent the second")
	}

	g.WithWriteLock(func() { g.resetActivationTurnTallyLocked() })
	if !onlyOnceSpent(g, src, onlyOnceProbe()) {
		t.Error("the turn's tally was emptied and the ability came back: it is once ever, not once each turn")
	}

	stealForTest(t, g, src, opp.ID)
	if !onlyOnceSpent(g, src, onlyOnceProbe()) {
		t.Error("the new controller reads the ability as unspent — CR 602.5b keeps the limit on the object")
	}
}

// TestAnOnlyOnceAbilityComesBackOnANewObject — CR 400.7: the permanent
// that returns is a new object, and nothing has activated its ability.
func TestAnOnlyOnceAbilityComesBackOnANewObject(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "only-once-flicker", onlyOnceProbe())
	src := pushExhaustSource(g, me, "only-once-flicker")

	if err := fireOnlyOnce(t, g, me, src, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passBothForTest(g)
	me.ManaPool = nil
	bounceAndReturn(t, g, src, me.ID)
	if onlyOnceSpent(g, src, onlyOnceProbe()) {
		t.Fatal("the returned permanent reads the old object's activation")
	}
	if err := fireOnlyOnce(t, g, me, src, 0); err != nil {
		t.Errorf("the returned permanent cannot activate it: %v", err)
	}
}

// TestAnOnlyOnceAbilityIsNotAnExhaustAbility — the announcement is not
// stamped as an exhaust one, so "whenever you activate an exhaust
// ability" does not see it, and the exhaust gate does not claim it.
func TestAnOnlyOnceAbilityIsNotAnExhaustAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "only-once-not-exhaust", onlyOnceProbe())
	src := pushExhaustSource(g, me, "only-once-not-exhaust")

	if err := fireOnlyOnce(t, g, me, src, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	announced := false
	for _, ev := range g.Events {
		if ev.Kind != EventActivateAbility || ev.CardID != src {
			continue
		}
		announced = true
		if ev.Exhaust {
			t.Error("the announcement is stamped Exhaust")
		}
	}
	if !announced {
		t.Fatal("no EventActivateAbility for the activation")
	}
	var n int
	var exhausted bool
	g.WithWriteLock(func() {
		n = g.ExhaustAbilitiesActivatedThisTurn(me.ID)
		exhausted = g.AbilityExhausted(me.ID, src, onlyOnceProbe())
	})
	if n != 0 {
		t.Errorf("ExhaustAbilitiesActivatedThisTurn = %d, want 0", n)
	}
	if exhausted {
		t.Error("AbilityExhausted claims an ability that is not an exhaust ability")
	}
}

// TestAnOnlyOnceAbilityStaysSpentAcrossASnapshotRoundTrip — the record
// is game state, so a restored game refuses the second activation as the
// live one does.
func TestAnOnlyOnceAbilityStaysSpentAcrossASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "only-once-snapshot", onlyOnceProbe())
	src := pushExhaustSource(g, me, "only-once-snapshot")
	if err := fireOnlyOnce(t, g, me, src, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passBothForTest(g)

	_, restored := roundTrip(t, g)
	if !onlyOnceSpent(restored, src, onlyOnceProbe()) {
		t.Fatal("the restored game reads the ability as unspent")
	}
	if err := fireOnlyOnce(t, restored, restored.Seats[0], src, 0); !errors.Is(err, ErrAbilityActivatedOnce) {
		t.Errorf("the restored game: err = %v, want ErrAbilityActivatedOnce", err)
	}
}

// TestOnlyOnceSpentIsFalseWithoutTheBit — the ordinary ability is never
// limited, however often it is activated.
func TestOnlyOnceSpentIsFalseWithoutTheBit(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	plain := onlyOnceProbe()
	plain.OnlyOnce = false
	exhaustCatalog(t, "only-once-plain", plain)
	src := pushExhaustSource(g, me, "only-once-plain")
	for i := 0; i < 2; i++ {
		if err := fireOnlyOnce(t, g, me, src, 0); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		passBothForTest(g)
	}
	if onlyOnceSpent(g, src, plain) {
		t.Error("OnlyOnceSpent = true for an ability without the bit")
	}
}
