package testsg

import (
	"testing"

	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// The headline: chapter I resolved means no discard at cleanup while the
// Saga is controlled, and the view reports no maximum.
func TestGreatSynthesisChapterOneLiftsTheMaximumWhileYouControlTheSaga(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := effectstest.GreatSynthesisToChapterOne(t, g, me)
	if got := effectstest.MhMax(g, me); got != 7 {
		t.Fatalf("max with chapter I still on the stack = %d, want 7 (the static is gone)", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != 14 {
		t.Fatalf("hand after chapter I = %d, want 14", got)
	}
	if got := effectstest.MhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("max after chapter I = %d, want no maximum", got)
	}
	for _, s := range protocol.ViewOfGameFor(g, me.ID.String()).Seats {
		if s.ID == me.ID.String() && s.MaxHandSize != game.NoMaxHandSize {
			t.Errorf("view max_hand_size = %d, want %d", s.MaxHandSize, game.NoMaxHandSize)
		}
	}
	effectstest.RunCleanup(t, g)
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
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := effectstest.GreatSynthesisToChapterOne(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MhMax(g, me); got != game.NoMaxHandSize {
		t.Fatalf("setup: max = %d, want no maximum", got)
	}

	effectstest.LeaveBattlefield(t, g, me.ID, saga)
	if got := effectstest.MhMax(g, me); got != 7 {
		t.Errorf("max after the Saga left = %d, want 7", got)
	}
	effectstest.RunCleanup(t, g)
	if n := g.DiscardPending[me.ID]; n != 7 {
		t.Errorf("cleanup asked for %d discards from a %d-card hand, want 7", n, me.Hand.Size())
	}
}

// Chapter III flickers the Saga: a new object (CR 400.7), so "this Saga" is
// gone and the effect ends with it.
func TestGreatSynthesisFlickerInChapterThreeEndsIt(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.GreatSynthesisToChapterOne(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	saga := effectstest.PraetorOnBattlefield(g, cardsj.JinGitaxiasOracleID)
	g.WithWriteLock(func() {
		saga.Counters[game.CounterLore] = 2
	})
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	// #4250: the free casts come first, inside the chapter's resolution;
	// declining them lets the flicker run.
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		effectstest.DeclineResolutionCast(t, g, me.ID)
	}

	if c := effectstest.PraetorOnBattlefield(g, cardsj.JinGitaxiasOracleID); c == nil || c.Name != "Jin-Gitaxias" {
		t.Fatalf("the Saga did not come back front face up: %+v", c)
	}
	if got := effectstest.MhMax(g, me); got != 7 {
		t.Errorf("max after chapter III's flicker = %d, want 7", got)
	}
}

// Countering or stifling chapter I leaves no effect at all (CR 701.6b: the
// ability ceases to exist), even though the Saga stays on the battlefield.
func TestGreatSynthesisCounteredChapterOneLeavesNoEffect(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := effectstest.GreatSynthesisToChapterOne(t, g, me)
	chapter := effectstest.TriggerOnStack(g, saga)
	var err error
	g.WithWriteLock(func() { err = g.CounterTargetForEffect(chapter.ID) })
	if err != nil {
		t.Fatalf("counter chapter I: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(saga) {
		t.Fatal("the Saga left the battlefield")
	}
	if got := me.Hand.Size(); got != 7 {
		t.Errorf("hand = %d, want 7 (chapter I never drew)", got)
	}
	if got := effectstest.MhMax(g, me); got != 7 {
		t.Errorf("max = %d, want 7: a countered chapter I leaves no effect", got)
	}
	effectstest.RunCleanup(t, g)
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
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := effectstest.GreatSynthesisToChapterOne(t, g, me)
	effectstest.LeaveBattlefield(t, g, me.ID, saga)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MhMax(g, me); got != 7 {
		t.Errorf("max = %d, want 7", got)
	}
	effectstest.RunCleanup(t, g)
	if n := g.DiscardPending[me.ID]; n != me.Hand.Size()-7 || n == 0 {
		t.Errorf("cleanup asked for %d discards from a %d-card hand, want %d", n, me.Hand.Size(), me.Hand.Size()-7)
	}
}

// A clone owns its own copy of the effect, and a clone is read the same way:
// ending the effect on the original does not reach it, and the grant carries
// its source-bound duration.
func TestGreatSynthesisEffectSurvivesCloneAndIsIndependent(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := effectstest.GreatSynthesisToChapterOne(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)

	clone := g.Clone()
	cloneMe := clone.PlayerByIDForEffect(me.ID)
	if got := effectstest.MhMax(clone, cloneMe); got != game.NoMaxHandSize {
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

	effectstest.LeaveBattlefield(t, g, me.ID, saga)
	if got := effectstest.MhMax(g, me); got != 7 {
		t.Errorf("original after the Saga left = %d, want 7", got)
	}
	if got := effectstest.MhMax(clone, cloneMe); got != game.NoMaxHandSize {
		t.Errorf("the clone follows the original: max = %d, want no maximum", got)
	}
}

// The grant is plain data: it survives a JSON snapshot with its source-bound
// duration, and the restored game still ends it when the Saga leaves.
func TestGreatSynthesisEffectSurvivesASnapshotRoundTrip(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := effectstest.GreatSynthesisToChapterOne(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)

	restored := effectstest.RestoreThroughJSON(t, g)
	rme := restored.PlayerByIDForEffect(me.ID)
	if got := effectstest.MhMax(restored, rme); got != game.NoMaxHandSize {
		t.Fatalf("restored max = %d, want no maximum", got)
	}
	effectstest.LeaveBattlefield(t, restored, me.ID, saga)
	if got := effectstest.MhMax(restored, rme); got != 7 {
		t.Errorf("restored max after the Saga left = %d, want 7", got)
	}
}
