package game

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// exile_until_another_test.go — #2539, ADR 0066's 2026-10-08
// amendment: "you may play it until you exile another card with this
// artifact". The window is an UntilSourceExilesAnother duration naming
// the holder and the source OBJECT; the next exile by the same linked
// ability for the same holder marks it Ended.

// untilAnotherItem is an activated ability of `source` (as object
// `epoch`) controlled by `controller`. SourceObjectForEffect reads the
// stamp, so the source need not be on the battlefield: that is the
// "sacrificed in response" case, and every test can use it.
func untilAnotherItem(controller, source uuid.UUID, epoch int) *StackItem {
	return &StackItem{
		Kind:         StackItemActivated,
		Controller:   controller,
		SourceCardID: source,
		SourceObject: ObjectRef{ID: source, Epoch: epoch},
	}
}

// exileUntilAnother runs the helper for `item` and returns the card it
// exiled from the controller's library top (uuid.Nil when none).
func exileUntilAnother(t *testing.T, g *Game, item *StackItem) uuid.UUID {
	t.Helper()
	var top uuid.UUID
	g.WithWriteLock(func() {
		if p := g.playerByIDLocked(item.Controller); p != nil && p.Library != nil && len(p.Library.Cards) > 0 {
			top = p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		}
		if err := g.ExileTopUntilAnotherForEffect(item, 1); err != nil {
			t.Fatalf("ExileTopUntilAnotherForEffect: %v", err)
		}
	})
	if top != uuid.Nil && !g.Exile.Contains(top) {
		t.Fatalf("the top card %v was not exiled", top)
	}
	return top
}

func TestExileUntilAnotherOpensAWindowNamingTheSourceObject(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	amulet := uuid.New()
	seedLibraryTop(me, "First", "Instant")

	first := exileUntilAnother(t, g, untilAnotherItem(me.ID, amulet, 3))
	if !castableFromExileBy(g, first, me.ID) {
		t.Fatal("the exiled card is not playable")
	}
	d := exilePlayOf(g, first).Duration
	want := Duration{Kind: UntilSourceExilesAnother, Player: me.ID, Source: amulet, SourceEpoch: 3}
	if !d.Equal(want) {
		t.Fatalf("duration = %+v, want %+v", d, want)
	}
	// No turn boundary ends it: this turn's cleanup, the next seat's
	// turn, and my own next turn.
	mySeat := g.Turn.ActiveSeat
	advanceToStepOfSeat(t, g, (mySeat+1)%4, StepPrecombatMain)
	advanceToStepOfSeat(t, g, mySeat, StepPrecombatMain)
	if !castableFromExileBy(g, first, me.ID) {
		t.Error("a turn boundary closed the window")
	}
}

func TestExileUntilAnotherClosesTheEarlierWindow(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	amulet := uuid.New()
	item := untilAnotherItem(me.ID, amulet, 0)

	seedLibraryTop(me, "First", "Instant")
	first := exileUntilAnother(t, g, item)
	seedLibraryTop(me, "Second", "Instant")
	second := exileUntilAnother(t, g, item)

	if castableFromExileBy(g, first, me.ID) {
		t.Error("the first card is still playable after another card was exiled with the source")
	}
	if !castableFromExileBy(g, second, me.ID) {
		t.Error("the newest card is not playable")
	}
	if !g.Exile.Contains(first) {
		t.Error("the first card left exile; it should stay there, unplayable")
	}
	// The ended window is hygiene for the next sweep.
	g.WithWriteLock(func() { g.sweepCastPermissionsLocked(false) })
	if n := len(me.CastPermissions); n != 1 {
		t.Errorf("after the sweep %d permissions remain, want the newest only", n)
	}
}

// CR 400.7: a NEW object of the same card is not "this artifact".
func TestExileUntilAnotherFromANewSourceObjectLeavesTheOldWindow(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	amulet := uuid.New()

	seedLibraryTop(me, "First", "Instant")
	first := exileUntilAnother(t, g, untilAnotherItem(me.ID, amulet, 1))
	seedLibraryTop(me, "Second", "Instant")
	second := exileUntilAnother(t, g, untilAnotherItem(me.ID, amulet, 2))

	if !castableFromExileBy(g, first, me.ID) {
		t.Error("an exile by the source's NEW object closed the old object's window")
	}
	if !castableFromExileBy(g, second, me.ID) {
		t.Error("the new object's card is not playable")
	}
}

// Owner answer 4: only the holder's own exile closes the window. An
// opponent who controls the same source object exiles a card "with
// this artifact", but "you" did not.
func TestExileUntilAnotherByAnotherPlayerLeavesTheHoldersWindow(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thief := g.Seats[(g.Turn.ActiveSeat+1)%4]
	toMainPhase(t, g)
	amulet := uuid.New()

	seedLibraryTop(me, "Mine", "Instant")
	mine := exileUntilAnother(t, g, untilAnotherItem(me.ID, amulet, 0))
	seedLibraryTop(thief, "Theirs", "Instant")
	theirs := exileUntilAnother(t, g, untilAnotherItem(thief.ID, amulet, 0))

	if !castableFromExileBy(g, mine, me.ID) {
		t.Error("another player's exile with the source closed my window")
	}
	if !castableFromExileBy(g, theirs, thief.ID) {
		t.Error("the new controller's card is not playable by them")
	}
	if castableFromExileBy(g, theirs, me.ID) {
		t.Error("the new controller's card is playable by me")
	}
}

// Nothing exiled, nothing closed: an empty library leaves the window.
func TestExileUntilAnotherWithAnEmptyLibraryKeepsTheWindow(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	amulet := uuid.New()
	item := untilAnotherItem(me.ID, amulet, 0)

	seedLibraryTop(me, "Only", "Instant")
	g.WithWriteLock(func() {
		// Leave just the card we are about to exile.
		me.Library.Cards = me.Library.Cards[len(me.Library.Cards)-1:]
	})
	only := exileUntilAnother(t, g, item)
	if got := exileUntilAnother(t, g, item); got != uuid.Nil {
		t.Fatalf("exiled %v from an empty library", got)
	}
	if !castableFromExileBy(g, only, me.ID) {
		t.Error("an exile that exiled nothing closed the window")
	}
}

// Playing the card moves it out of exile, which ends the permission by
// object identity (ADR 0066 decision 2); a land is played under the
// usual land-play rules (CR 305.1).
func TestExileUntilAnotherLetsALandBePlayedOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	seedLibraryTop(me, "Exiled Mountain", "Basic Land — Mountain")
	land := exileUntilAnother(t, g, untilAnotherItem(me.ID, uuid.New(), 0))
	if !castableFromExileBy(g, land, me.ID) {
		t.Fatal("a land exiled this way is not playable")
	}
	if err := g.CastSpell(me.ID, land, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the land from exile: %v", err)
	}
	if findCardOnBattlefield(g, land) < 0 {
		t.Fatal("the land is not on the battlefield")
	}
	g.WithWriteLock(func() { g.sweepCastPermissionsLocked(false) })
	if n := len(me.CastPermissions); n != 0 {
		t.Errorf("%d permissions outlived the card leaving exile", n)
	}
}

func TestExileUntilAnotherSurvivesCloneAndSnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	item := untilAnotherItem(me.ID, uuid.New(), 4)

	seedLibraryTop(me, "First", "Instant")
	first := exileUntilAnother(t, g, item)
	before := g.Clone()
	seedLibraryTop(me, "Second", "Instant")
	second := exileUntilAnother(t, g, item)

	// The undo snapshot taken before the second exile still has the
	// first window open: closing it wrote a fresh slice.
	if !castableFromExileBy(before, first, before.Seats[g.Turn.ActiveSeat].ID) {
		t.Error("closing the window rewrote an earlier clone")
	}

	for _, tc := range []struct {
		name string
		make func(t *testing.T) *Game
	}{
		{"clone", func(t *testing.T) *Game { return g.Clone() }},
		{"snapshot round-trip", func(t *testing.T) *Game {
			restored, err := g.CaptureSnapshot().Restore()
			if err != nil {
				t.Fatalf("Restore: %v", err)
			}
			return restored
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.make(t)
			if castableFromExileBy(out, first, me.ID) {
				t.Error("the ended window came back open")
			}
			if !castableFromExileBy(out, second, me.ID) {
				t.Error("the open window came back closed")
			}
			// And the restored source still closes it.
			seedLibraryTop(out.Seats[out.Turn.ActiveSeat], "Third", "Instant")
			third := exileUntilAnother(t, out, item)
			if castableFromExileBy(out, second, me.ID) || !castableFromExileBy(out, third, me.ID) {
				t.Error("the restored window is not closed by the next exile")
			}
		})
	}
}

func TestUntilSourceExilesAnotherProblem(t *testing.T) {
	src := uuid.New()
	for _, tc := range []struct {
		name string
		d    Duration
		want string
	}{
		{"open", Duration{Kind: UntilSourceExilesAnother, Source: src, SourceEpoch: 2}, ""},
		{"ended", Duration{Kind: UntilSourceExilesAnother, Source: src, Ended: true}, ""},
		{"no source", Duration{Kind: UntilSourceExilesAnother}, "names no source"},
		{"epoch on another kind", Duration{Kind: Indefinite, SourceEpoch: 1}, "source-exile stamp"},
		{"ended on another kind", Duration{Kind: WhileInZone, Ended: true}, "source-exile stamp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.d.Problem()
			if (tc.want == "") != (got == "") || !strings.Contains(got, tc.want) {
				t.Errorf("Problem() = %q, want %q", got, tc.want)
			}
		})
	}
}
