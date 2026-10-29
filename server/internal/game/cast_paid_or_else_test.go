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

var paidOrElseCount = RegisterCastFollowUp(testEffectKeyPrefix+"paid-or-else-count", func(_ *Game, f CastFollowUp) error {
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
// runs the offer from a fresh source, inside the resolution of an
// ability (CR 608.2g, #4250), so the cast is offered as a resolution_cast
// that holds that resolution open.
func paidOrElseTable(t *testing.T, top *Card) (*Game, *Player, uuid.UUID) {
	return paidOrElseTableAt(t, top, true)
}

// paidOrElseTableAt is paidOrElseTable, with the offer made outside any
// resolution when inResolution is false: the fallback, where the
// pass-closed permission alone is the window.
func paidOrElseTableAt(t *testing.T, top *Card, inResolution bool) (*Game, *Player, uuid.UUID) {
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
	offer := func(g *Game) error {
		return g.ExileTopThenPaidCastOrElseForEffect(PaidCastOrElse{
			Player: me.ID, Source: source, SourceName: "Test Source", IfNotCast: paidOrElseCount,
			DeclineLabel: "Don't cast it",
		})
	}
	if inResolution {
		pushAbilityItem(g, StackItemActivated, me.ID, uuid.Nil, "Fixture paid-or-else", func(g *Game, _ *StackItem) error {
			return offer(g)
		})
		resolveTop(t, g)
		return g, me, source
	}
	g.WithWriteLock(func() {
		if err := offer(g); err != nil {
			t.Fatalf("ExileTopThenPaidCastOrElseForEffect: %v", err)
		}
	})
	return g, me, source
}

// The offer is a resolution_cast over the exiled card, holding the
// resolution open; nothing is decided before the answer, and declining
// runs "if you don't" once, inside the resolution, with the card left in
// exile.
func TestPaidCastOrElseAsksOverTheExiledCardAndDeclineRunsIfNotCast(t *testing.T) {
	top := libCard(uuid.Nil, "Divination", "Sorcery", "{2}{U}")
	g, me, source := paidOrElseTable(t, &top)
	c := resolutionCastPrompt(g, me.ID)
	if c == nil || len(c.DiscardOptions) != 1 || c.DiscardOptions[0] != top.InstanceID || c.MayCastKeyword != MayCastKeywordPaidOrElse {
		t.Fatalf("prompt = %+v, want a paid_or_else resolution_cast over the exiled card", c)
	}
	if c.DeclineLabel != "Don't cast it" || c.Reason != "Test Source" {
		t.Errorf("prompt words = %q / %q", c.Reason, c.DeclineLabel)
	}
	if !g.ResolutionPaused() {
		t.Fatal("the offer is not holding the resolution open")
	}
	if paidOrElseRunsFor(source) != 0 {
		t.Fatal("nothing is decided before the answer")
	}
	answerResolutionCast(t, g, me.ID, false)
	if got := paidOrElseRunsFor(source); got != 1 {
		t.Errorf("decline ran if-not-cast %d times, want 1", got)
	}
	if !g.Exile.Contains(top.InstanceID) || g.CastPermissionOnCardByIDForEffect(top.InstanceID) != nil {
		t.Error("declined: the card stays in exile with no grant")
	}
	if g.ResolutionPaused() {
		t.Error("the decline did not finish the resolution")
	}
}

// Under the prompt the card carries the paid, any-timing, pass-closed
// grant: the printed cost, flash timing, cast only, LapseStaysInExile
// and the "if you don't" body.
func TestPaidCastOrElsePromptCarriesAPaidPassClosedGrant(t *testing.T) {
	top := libCard(uuid.Nil, "Divination", "Sorcery", "{2}{U}")
	g, me, source := paidOrElseTable(t, &top)
	perm := g.CastPermissionOnCardByIDForEffect(top.InstanceID)
	if perm == nil || perm.Cost != "" || perm.Timing != TimingFlash || !perm.CastOnly ||
		perm.LapseOnPass != LapseStaysInExile || perm.IfNotCast != paidOrElseCount || perm.Player != me.ID {
		t.Fatalf("grant = %+v", perm)
	}
	if paidOrElseRunsFor(source) != 0 {
		t.Error("the offer is not yet a decision")
	}
}

// Outside a resolution there is nothing to hold open: the grant alone is
// the window, and passing without casting runs "if you don't" once.
func TestPaidCastOrElsePassingWithoutCastingRunsIfNotCastOnce(t *testing.T) {
	top := libCard(uuid.Nil, "Divination", "Sorcery", "{2}{U}")
	g, me, source := paidOrElseTableAt(t, &top, false)
	if resolutionCastPrompt(g, me.ID) != nil {
		t.Fatal("no resolution is open, so nothing is asked")
	}
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

// Casting through the prompt, paying the card's cost, runs nothing, and
// the spell is on the stack before the resolution finishes.
func TestPaidCastOrElseCastingRunsNothing(t *testing.T) {
	top := libCard(uuid.Nil, "Shock", "Instant", "{R}")
	g, me, source := paidOrElseTable(t, &top)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	answerResolutionCast(t, g, me.ID, true)
	if !g.Stack.Contains(top.InstanceID) || g.ResolutionPaused() {
		t.Fatal("the paid cast is not on the stack, or the resolution is still held")
	}
	for i := 0; i < 4; i++ {
		_ = g.PassPriority()
	}
	if got := paidOrElseRunsFor(source); got != 0 {
		t.Errorf("a cast card ran if-not-cast %d times", got)
	}
}

// The prompt's cast is strict: a card its caster cannot pay for is
// refused and the prompt stays open for the decline.
func TestPaidCastOrElseUnpayableCastLeavesThePromptOpen(t *testing.T) {
	top := libCard(uuid.Nil, "Divination", "Sorcery", "{2}{U}")
	g, me, source := paidOrElseTable(t, &top)
	c := resolutionCastPrompt(g, me.ID)
	if err := g.CastForResolutionChoice(me.ID, c.ID, top.InstanceID, CastSpellParams{FromZone: string(ZoneExile)}); err == nil {
		t.Fatal("an unpaid cast was accepted")
	}
	if resolutionCastPrompt(g, me.ID) == nil || !g.ResolutionPaused() || paidOrElseRunsFor(source) != 0 {
		t.Fatal("the refused cast closed the prompt or ran if-not-cast")
	}
	answerResolutionCast(t, g, me.ID, false)
	if got := paidOrElseRunsFor(source); got != 1 {
		t.Errorf("decline after a refused cast ran if-not-cast %d times, want 1", got)
	}
}

// A table paused on the offer is a restore point: the restored prompt
// still runs "if you don't" on a decline, and casts on a cast.
func TestPaidCastOrElseOfferSurvivesARestoreMidPrompt(t *testing.T) {
	top := libCard(uuid.Nil, "Divination", "Sorcery", "{2}{U}")
	g, _, source := paidOrElseTable(t, &top)
	_, restored := roundTrip(t, g)
	rme := restored.Seats[restored.Turn.PriorityHolder]
	if rc := resolutionCastPrompt(restored, rme.ID); rc == nil || !restored.ResolutionPaused() ||
		rc.MayCastKeyword != MayCastKeywordPaidOrElse {
		t.Fatalf("the restored table lost the offer: %+v", rc)
	}
	before := paidOrElseRunsFor(source)
	answerResolutionCast(t, restored, rme.ID, false)
	if got := paidOrElseRunsFor(source) - before; got != 1 {
		t.Errorf("the restored decline ran if-not-cast %d times, want 1", got)
	}
	if restored.ResolutionPaused() {
		t.Error("the restored resolution did not finish")
	}
}

func TestPaidCastOrElseSweepBackstopRunsIfNotCast(t *testing.T) {
	top := libCard(uuid.Nil, "Divination", "Sorcery", "{2}{U}")
	g, _, source := paidOrElseTableAt(t, &top, false)
	g.WithWriteLock(func() { g.sweepCastPermissionsLocked(true) })
	if got := paidOrElseRunsFor(source); got != 1 {
		t.Errorf("the end-of-turn sweep ran if-not-cast %d times, want 1", got)
	}
}

func TestPaidCastOrElseLandAndEmptyLibraryRunIfNotCastWithNoQuestion(t *testing.T) {
	land := libCard(uuid.Nil, "Mountain", "Basic Land — Mountain", "")
	g, me, source := paidOrElseTable(t, &land)
	if resolutionCastPrompt(g, me.ID) != nil {
		t.Fatal("a land can't be cast; nothing is asked")
	}
	if got := paidOrElseRunsFor(source); got != 1 || !g.Exile.Contains(land.InstanceID) {
		t.Errorf("land: ran %d times, in exile %v", got, g.Exile.Contains(land.InstanceID))
	}

	_, _, empty := paidOrElseTable(t, nil)
	if got := paidOrElseRunsFor(empty); got != 1 {
		t.Errorf("empty library: ran %d times, want 1", got)
	}
}
