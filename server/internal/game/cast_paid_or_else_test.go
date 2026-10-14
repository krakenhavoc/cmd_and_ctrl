package game

import (
	"sync"
	"testing"

	"github.com/google/uuid"
)

// cast_paid_or_else_test.go — #3510. "Exile the top card of your library.
// You may cast that card. If you don't, ..." The "if you don't" body runs
// exactly once on every path where the card is not cast, and never when
// it is.

var (
	paidOrElseMu   sync.Mutex
	paidOrElseRuns = map[uuid.UUID]int{}
)

var paidOrElseCount = RegisterCastFollowUp("test-paid-or-else-count", func(_ *Game, f CastFollowUp) error {
	paidOrElseMu.Lock()
	paidOrElseRuns[f.Source]++
	paidOrElseMu.Unlock()
	return nil
})

func paidOrElseRunsFor(source uuid.UUID) int {
	paidOrElseMu.Lock()
	defer paidOrElseMu.Unlock()
	return paidOrElseRuns[source]
}

// paidOrElseTable puts `top` on top of the priority holder's library and
// runs the offer from a fresh source.
func paidOrElseTable(t *testing.T, top *Card) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[g.Turn.PriorityHolder]
	if top != nil {
		top.Owner, top.Controller = me.ID, me.ID
		libraryOf(t, me, *top)
	} else {
		me.Library.Cards = nil
	}
	source := uuid.New()
	g.WithWriteLock(func() {
		if err := g.ExileTopThenPaidCastOrElseForEffect(PaidCastOrElse{
			Player: me.ID, Source: source, SourceName: "Test Source", IfNotCast: paidOrElseCount,
		}); err != nil {
			t.Fatalf("ExileTopThenPaidCastOrElseForEffect: %v", err)
		}
	})
	return g, me, source
}

func TestPaidCastOrElseAsksOverTheExiledCardAndDeclineRunsIfNotCast(t *testing.T) {
	top := libCard(uuid.Nil, "Divination", "Sorcery", "{2}{U}")
	g, _, source := paidOrElseTable(t, &top)
	c := discoverPrompt(t, g)
	if c.MayCastCard != top.InstanceID || c.MayCastKeyword != MayCastKeywordPaidOrElse {
		t.Fatalf("prompt = card %s keyword %q", c.MayCastCard, c.MayCastKeyword)
	}
	if paidOrElseRunsFor(source) != 0 {
		t.Fatal("nothing is decided before the answer")
	}
	answerMayCast(t, g, c.Chooser, false)
	if got := paidOrElseRunsFor(source); got != 1 {
		t.Errorf("decline ran if-not-cast %d times, want 1", got)
	}
	if !g.Exile.Contains(top.InstanceID) || g.CastPermissionOnCardByIDForEffect(top.InstanceID) != nil {
		t.Error("declined: the card stays in exile with no grant")
	}
}

func TestPaidCastOrElseAcceptGrantsAPaidPassClosedCast(t *testing.T) {
	top := libCard(uuid.Nil, "Divination", "Sorcery", "{2}{U}")
	g, me, source := paidOrElseTable(t, &top)
	answerMayCast(t, g, me.ID, true)
	perm := g.CastPermissionOnCardByIDForEffect(top.InstanceID)
	if perm == nil || perm.Cost != "" || perm.Timing != TimingFlash || !perm.CastOnly ||
		perm.LapseOnPass != LapseStaysInExile || perm.IfNotCast != paidOrElseCount || perm.Player != me.ID {
		t.Fatalf("grant = %+v", perm)
	}
	if paidOrElseRunsFor(source) != 0 {
		t.Error("accepting is not yet a decision")
	}
}

func TestPaidCastOrElsePassingWithoutCastingRunsIfNotCastOnce(t *testing.T) {
	top := libCard(uuid.Nil, "Divination", "Sorcery", "{2}{U}")
	g, me, source := paidOrElseTable(t, &top)
	answerMayCast(t, g, me.ID, true)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if got := paidOrElseRunsFor(source); got != 1 {
		t.Fatalf("the lapse ran if-not-cast %d times, want 1", got)
	}
	if !g.Exile.Contains(top.InstanceID) || g.CastPermissionOnCardByIDForEffect(top.InstanceID) != nil {
		t.Error("lapsed: the card stays in exile and the window is closed")
	}
	for i := 0; i < 3; i++ {
		_ = g.PassPriority()
	}
	if got := paidOrElseRunsFor(source); got != 1 {
		t.Errorf("later passes ran it again: %d", got)
	}
}

func TestPaidCastOrElseCastingRunsNothing(t *testing.T) {
	top := libCard(uuid.Nil, "Shock", "Instant", "{R}")
	g, me, source := paidOrElseTable(t, &top)
	answerMayCast(t, g, me.ID, true)
	if err := g.CastSpell(me.ID, top.InstanceID, CastSpellParams{FromZone: string(ZoneExile)}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	for i := 0; i < 4; i++ {
		_ = g.PassPriority()
	}
	if got := paidOrElseRunsFor(source); got != 0 {
		t.Errorf("a cast card ran if-not-cast %d times", got)
	}
}

func TestPaidCastOrElseSweepBackstopRunsIfNotCast(t *testing.T) {
	top := libCard(uuid.Nil, "Divination", "Sorcery", "{2}{U}")
	g, me, source := paidOrElseTable(t, &top)
	answerMayCast(t, g, me.ID, true)
	g.WithWriteLock(func() { g.sweepCastPermissionsLocked(true) })
	if got := paidOrElseRunsFor(source); got != 1 {
		t.Errorf("the end-of-turn sweep ran if-not-cast %d times, want 1", got)
	}
}

func TestPaidCastOrElseLandAndEmptyLibraryRunIfNotCastWithNoQuestion(t *testing.T) {
	land := libCard(uuid.Nil, "Mountain", "Basic Land — Mountain", "")
	g, _, source := paidOrElseTable(t, &land)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMayCast {
			t.Fatal("a land can't be cast; nothing is asked")
		}
	}
	if got := paidOrElseRunsFor(source); got != 1 || !g.Exile.Contains(land.InstanceID) {
		t.Errorf("land: ran %d times, in exile %v", got, g.Exile.Contains(land.InstanceID))
	}

	_, _, empty := paidOrElseTable(t, nil)
	if got := paidOrElseRunsFor(empty); got != 1 {
		t.Errorf("empty library: ran %d times, want 1", got)
	}
}
