package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// great_synthesis_hand_size_test.go — The Great Synthesis chapter I's "you
// have no maximum hand size for as long as you control this Saga" as an
// effect the resolving ability CREATES (ADR 0113's amendment of 2026-10-13,
// #3247), not a static of the Saga.

// greatSynthesisToChapterOne flips Jin-Gitaxias with seven cards in hand and
// stops with the Saga on the battlefield and its chapter I trigger on the
// stack. Returns the Saga's instance ID.
func greatSynthesisToChapterOne(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	jin := pushBattlefieldCardWithTimestamp(g, jinCard(me.ID, false))
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	praetorFillHand(me, 7)
	toMainForCost(t, g)
	floatForTest(g, me, "UUUU")
	if err := g.ActivateCatalogAbility(me.ID, jin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	for i := 0; i < 12; i++ {
		if saga := praetorOnBattlefield(g, jinGitaxiasOracleID); saga != nil && saga.Name == "The Great Synthesis" &&
			triggerOnStack(g, saga.InstanceID) != nil {
			return saga.InstanceID
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("chapter I never reached the stack")
	return uuid.Nil
}

// The headline: chapter I resolved means no discard at cleanup while the
// Saga is controlled, and the view reports no maximum.
func TestGreatSynthesisChapterOneLiftsTheMaximumWhileYouControlTheSaga(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := greatSynthesisToChapterOne(t, g, me)
	if got := mhMax(g, me); got != 7 {
		t.Fatalf("max with chapter I still on the stack = %d, want 7 (the static is gone)", got)
	}
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != 14 {
		t.Fatalf("hand after chapter I = %d, want 14", got)
	}
	if got := mhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("max after chapter I = %d, want no maximum", got)
	}
	for _, s := range protocol.ViewOfGameFor(g, me.ID.String()).Seats {
		if s.ID == me.ID.String() && s.MaxHandSize != game.NoMaxHandSize {
			t.Errorf("view max_hand_size = %d, want %d", s.MaxHandSize, game.NoMaxHandSize)
		}
	}
	runCleanup(t, g)
	if n := g.DiscardPending[me.ID]; n != 0 {
		t.Errorf("cleanup asked for %d discards with the Saga out", n)
	}
	if got := me.Hand.Size(); got != 14 {
		t.Errorf("hand at cleanup = %d, want 14", got)
	}
	if !g.Battlefield.Contains(saga) {
		t.Error("the Saga left the battlefield")
	}
}

// The Saga leaving ends it: the maximum is seven again, and a cleanup
// discards down to it.
func TestGreatSynthesisLeavingTheBattlefieldRestoresTheMaximum(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := greatSynthesisToChapterOne(t, g, me)
	passPriorityAroundTable(t, g)
	if got := mhMax(g, me); got != game.NoMaxHandSize {
		t.Fatalf("setup: max = %d, want no maximum", got)
	}

	leaveBattlefield(t, g, me.ID, saga)
	if got := mhMax(g, me); got != 7 {
		t.Errorf("max after the Saga left = %d, want 7", got)
	}
	runCleanup(t, g)
	if n := g.DiscardPending[me.ID]; n != 7 {
		t.Errorf("cleanup asked for %d discards from a %d-card hand, want 7", n, me.Hand.Size())
	}
}

// Chapter III flickers the Saga: a new object (CR 400.7), so "this Saga" is
// gone and the effect ends with it.
func TestGreatSynthesisFlickerInChapterThreeEndsIt(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	greatSynthesisToChapterOne(t, g, me)
	passPriorityAroundTable(t, g)
	saga := praetorOnBattlefield(g, jinGitaxiasOracleID)
	g.WithWriteLock(func() {
		saga.Counters[game.CounterLore] = 2
	})
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	if c := praetorOnBattlefield(g, jinGitaxiasOracleID); c == nil || c.Name != "Jin-Gitaxias" {
		t.Fatalf("the Saga did not come back front face up: %+v", c)
	}
	if got := mhMax(g, me); got != 7 {
		t.Errorf("max after chapter III's flicker = %d, want 7", got)
	}
}

// Countering or stifling chapter I leaves no effect at all (CR 701.6b: the
// ability ceases to exist), even though the Saga stays on the battlefield.
func TestGreatSynthesisCounteredChapterOneLeavesNoEffect(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := greatSynthesisToChapterOne(t, g, me)
	chapter := triggerOnStack(g, saga)
	var err error
	g.WithWriteLock(func() { err = g.CounterTargetForEffect(chapter.ID) })
	if err != nil {
		t.Fatalf("counter chapter I: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(saga) {
		t.Fatal("the Saga left the battlefield")
	}
	if got := me.Hand.Size(); got != 7 {
		t.Errorf("hand = %d, want 7 (chapter I never drew)", got)
	}
	if got := mhMax(g, me); got != 7 {
		t.Errorf("max = %d, want 7: a countered chapter I leaves no effect", got)
	}
	runCleanup(t, g)
	if n := g.DiscardPending[me.ID]; n != 0 {
		// Seven in hand against a maximum of seven: nothing to discard,
		// and the proof that the maximum was never lifted is mhMax above.
		t.Errorf("cleanup asked for %d discards", n)
	}
}

// A Saga that left in response to chapter I starts no effect (CR 611.2b: a
// duration already false as it would begin never begins). The draw still
// happens, and the cleanup discards.
func TestGreatSynthesisChapterOneAfterTheSagaLeftStartsNothing(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := greatSynthesisToChapterOne(t, g, me)
	leaveBattlefield(t, g, me.ID, saga)
	passPriorityAroundTable(t, g)

	if got := mhMax(g, me); got != 7 {
		t.Errorf("max = %d, want 7", got)
	}
	runCleanup(t, g)
	if n := g.DiscardPending[me.ID]; n != me.Hand.Size()-7 || n == 0 {
		t.Errorf("cleanup asked for %d discards from a %d-card hand, want %d", n, me.Hand.Size(), me.Hand.Size()-7)
	}
}

// A clone owns its own copy of the effect, and a clone is read the same way:
// ending the effect on the original does not reach it, and the grant carries
// its source-bound duration.
func TestGreatSynthesisEffectSurvivesCloneAndIsIndependent(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := greatSynthesisToChapterOne(t, g, me)
	passPriorityAroundTable(t, g)

	clone := g.Clone()
	cloneMe := clone.PlayerByIDForEffect(me.ID)
	if got := mhMax(clone, cloneMe); got != game.NoMaxHandSize {
		t.Fatalf("clone max = %d, want no maximum", got)
	}
	var found bool
	for _, s := range cloneMe.Statics {
		if s.HandSize.Active {
			found = true
			if s.Duration.Kind != game.ForAsLongAs || s.Duration.Condition != game.WhileYouControlSource || s.Duration.Source != saga {
				t.Errorf("cloned grant duration = %+v, want ForAsLongAs/WhileYouControlSource on the Saga", s.Duration)
			}
		}
	}
	if !found {
		t.Fatal("the clone lost the hand-size grant")
	}

	leaveBattlefield(t, g, me.ID, saga)
	if got := mhMax(g, me); got != 7 {
		t.Errorf("original after the Saga left = %d, want 7", got)
	}
	if got := mhMax(clone, cloneMe); got != game.NoMaxHandSize {
		t.Errorf("the clone follows the original: max = %d, want no maximum", got)
	}
}

// The grant is plain data: it survives a JSON snapshot with its source-bound
// duration, and the restored game still ends it when the Saga leaves.
func TestGreatSynthesisEffectSurvivesASnapshotRoundTrip(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := greatSynthesisToChapterOne(t, g, me)
	passPriorityAroundTable(t, g)

	restored := restoreThroughJSON(t, g)
	rme := restored.PlayerByIDForEffect(me.ID)
	if got := mhMax(restored, rme); got != game.NoMaxHandSize {
		t.Fatalf("restored max = %d, want no maximum", got)
	}
	leaveBattlefield(t, restored, me.ID, saga)
	if got := mhMax(restored, rme); got != 7 {
		t.Errorf("restored max after the Saga left = %d, want 7", got)
	}
}
