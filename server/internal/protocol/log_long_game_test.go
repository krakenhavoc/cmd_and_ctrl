package protocol

// log_long_game_test.go holds #3131's two claims about the per-frame
// cost of the log and the reveal window in a LONG game:
//
//   - the reveal window (publicRevealsOf) re-scanned every event the
//     game had ever emitted on every view; it now resumes a fold the
//     way the public log does, so a view pays for the events emitted
//     since the previous one;
//   - what is cached is exactly what a fresh projection would give,
//     after an undo, after a reveal to one player, and for a spectator
//     as against a seated player.

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// longGameEvents is how many events the benchmarks' game has emitted
// before the clock starts: a few hundred turns of a four-player
// commander game, an order of magnitude past the 200-entry log ring.
const longGameEvents = 30000

// growLongGame emits `n` events into g in the mix a real game
// produces, with a step every 40 events and a reveal every 150, so the
// reveal window always has a current-turn group to project.
func growLongGame(tb testing.TB, g *game.Game, n int) {
	tb.Helper()
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			active := g.Seats[(i/40)%len(g.Seats)]
			switch {
			case i%40 == 0:
				g.EmitEvent(game.Event{
					Kind: game.EventStepBegan, Actor: active.ID, Amount: i/40 + 1,
					Label: string(game.StepPrecombatMain),
				})
			case i%150 == 0:
				g.RevealTopOfLibraryForEffect(active.ID, uuid.Nil, RevealCardsMax, "bench")
			default:
				g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: active.ID})
			}
		}
		// End on the game's own turn with a full window of reveals, so
		// the projection has something to resolve every frame.
		revealer := g.Seats[0]
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: revealer.ID,
			Amount: g.Turn.Seq, Label: string(g.Turn.Step),
		})
		for i := 0; i < PublicRevealMax; i++ {
			g.RevealTopOfLibraryForEffect(revealer.ID, uuid.Nil, RevealCardsMax, "bench")
		}
	})
}

// BenchmarkLogAndRevealsLongGame is the per-frame cost of everything
// this issue touches that is not the board: both projections over a
// game that has emitted longGameEvents events, one new event per
// frame, ring full, the reveal window populated.
func BenchmarkLogAndRevealsLongGame(b *testing.B) {
	g := buildFourPlayerBoardB(b)
	growLongGame(b, g, longGameEvents)
	v := ViewOfGame(g)
	if len(v.Log) != PublicLogMax {
		b.Fatalf("log is not full: %d entries", len(v.Log))
	}
	if len(v.Reveals) == 0 {
		b.Fatal("the reveal window is empty")
	}
	actor := g.Seats[0].ID
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: actor})
		})
		_ = publicLogOf(g, &v)
		_ = publicRevealsOf(g, &v)
	}
}

// BenchmarkPublicRevealsLongGame is the reveal half on its own, with no
// new events: the steady frame the cache turns into a copy of a few
// groups.
func BenchmarkPublicRevealsLongGame(b *testing.B) {
	g := buildFourPlayerBoardB(b)
	growLongGame(b, g, longGameEvents)
	v := ViewOfGame(g)
	if len(v.Reveals) == 0 {
		b.Fatal("the reveal window is empty")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = publicRevealsOf(g, &v)
	}
}

// publicRevealsUncached is the reference: a fresh fold over every
// event, windowed against the same turn. Caller must hold g's read lock.
func publicRevealsUncached(g *game.Game, v *GameView) []RevealView {
	seatIDs := seatIDsOf(v)
	f := newRevealFold(seatIDs, g.EventLogGeneration())
	f.feed(g.Events)
	out := f.window(g.Turn.Seq)
	if len(out) == 0 {
		return nil
	}
	resolveRevealCards(out, v)
	return out
}

// publicRevealsReference is the pre-#3131 body of publicRevealsOf,
// kept verbatim so the fold is checked against the code it replaced
// rather than against itself.
func publicRevealsReference(g *game.Game, v *GameView) []RevealView {
	seatOf := seatIndexer(v)
	turn := 0
	var groups []RevealView
	index := make(map[uint64]int)
	for _, ev := range g.Events {
		if ev.Kind == game.EventStepBegan {
			turn = ev.Amount
			continue
		}
		if ev.Kind != game.EventRevealCards || ev.RevealSeq == 0 {
			continue
		}
		at, ok := index[ev.RevealSeq]
		if !ok {
			groups = append(groups, RevealView{
				Seq:      ev.RevealSeq,
				Turn:     turn,
				Seat:     seatOf(ev.Actor),
				Reason:   ev.Label,
				From:     string(ev.OldZone),
				To:       string(ev.RevealTo),
				sourceID: uuidStringOrEmpty(ev.Source),
			})
			at = len(groups) - 1
			index[ev.RevealSeq] = at
		}
		groups[at].cardIDs = append(groups[at].cardIDs, uuidStringOrEmpty(ev.CardID))
	}
	out := make([]RevealView, 0, PublicRevealMax)
	for i := len(groups) - 1; i >= 0 && len(out) < PublicRevealMax; i-- {
		if groups[i].Turn != g.Turn.Seq {
			break
		}
		out = append(out, groups[i])
	}
	if len(out) == 0 {
		return nil
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	resolveRevealCards(out, v)
	return out
}

// viewAndCheckReveals builds a view the production way (both cached
// folds) and fails unless its reveals — raw, and filtered for every seat
// and for a spectator — equal the pre-#3131 body's, and its log equals a
// fresh fold's. It returns the view.
func viewAndCheckReveals(t *testing.T, g *game.Game, where string) GameView {
	t.Helper()
	v := viewAndCheck(t, g, where)
	var ref []RevealView
	g.ReadSnapshot(func() { ref = publicRevealsReference(g, &v) })
	if !reflect.DeepEqual(v.Reveals, ref) {
		t.Fatalf("%s: cached reveals differ from the reference\ncached: %+v\nfresh:  %+v", where, v.Reveals, ref)
	}
	var fresh []RevealView
	g.ReadSnapshot(func() { fresh = publicRevealsUncached(g, &v) })
	if !reflect.DeepEqual(v.Reveals, fresh) {
		t.Fatalf("%s: cached reveals differ from a fresh fold\ncached: %+v\nfresh:  %+v", where, v.Reveals, fresh)
	}
	refView := v
	refView.Reveals = ref
	viewers := []string{spectatorID}
	for _, s := range v.Seats {
		viewers = append(viewers, s.ID)
	}
	for _, viewer := range viewers {
		got, want := FilterViewFor(v, viewer).Reveals, FilterViewFor(refView, viewer).Reveals
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: viewer %s: filtered cached reveals differ from the reference\ncached: %+v\nfresh:  %+v",
				where, viewer, got, want)
		}
	}
	return v
}

// TestCachedRevealsMatchAFreshProjection walks one game through the
// things that can make a resumed fold wrong: growth, a reveal to one
// player, an undo (the game is replaced by a clone and the log is
// restored to a shorter history that then regrows with the same Seq
// values), and a new turn that retires the window.
func TestCachedRevealsMatchAFreshProjection(t *testing.T) {
	g := buildFourPlayerBoard(t)
	revealer := g.Seats[0]

	viewAndCheckReveals(t, g, "initial")

	// A reveal to the whole table, then more events, viewed in between so
	// the fold is resumed rather than rebuilt.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: revealer.ID,
			Amount: g.Turn.Seq, Label: string(g.Turn.Step),
		})
		g.RevealTopOfLibraryForEffect(revealer.ID, uuid.Nil, 3, "first")
	})
	v := viewAndCheckReveals(t, g, "after a reveal")
	if len(v.Reveals) == 0 {
		t.Fatal("setup: no reveal reached the window")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: revealer.ID})
		g.RevealTopOfLibraryForEffect(g.Seats[1].ID, uuid.Nil, 2, "second")
	})
	v = viewAndCheckReveals(t, g, "after a second reveal")
	if len(v.Reveals) < 2 {
		t.Fatalf("want both reveals in the window, got %d", len(v.Reveals))
	}

	// A hand reveal: the window is built from the unfiltered view and
	// carries no handle, so every viewer gets the same window. Cached and
	// fresh must agree for each viewer.
	g.WithWriteLock(func() { g.RevealHandForEffect(g.Seats[2].ID) })
	viewAndCheckReveals(t, g, "after a hand reveal")

	// Undo: snapshot, grow past it, view (caching the longer history),
	// restore, then regrow to the SAME length with the same Seq values
	// and view again. Length and last Seq cannot tell the two histories
	// apart; only the event-log generation can.
	snap := g.Clone()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: revealer.ID})
		g.RevealTopOfLibraryForEffect(revealer.ID, uuid.Nil, 1, "undone")
	})
	viewAndCheckReveals(t, g, "before undo")
	// No view between the restore and the regrowth: a view of the shorter
	// log would rebuild the fold on its own and hide the case.
	g.WithWriteLock(func() {
		g.RestoreFrom(snap)
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: revealer.ID})
		g.RevealTopOfLibraryForEffect(revealer.ID, uuid.Nil, 1, "regrown")
	})
	v = viewAndCheckReveals(t, g, "after regrowth with reused Seq values")
	for _, r := range v.Reveals {
		if r.Reason == "undone" {
			t.Fatalf("the window still carries a reveal that was undone: %+v", r)
		}
	}

	// A new turn retires the window.
	g.WithWriteLock(func() {
		g.Turn.Seq++
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: g.Seats[1].ID,
			Amount: g.Turn.Seq, Label: string(g.Turn.Step),
		})
	})
	v = viewAndCheckReveals(t, g, "next turn")
	if len(v.Reveals) != 0 {
		t.Fatalf("a new turn should empty the window, got %d", len(v.Reveals))
	}
}

// TestCachedLogFollowsWhoKnowsTheCard holds the other half of the
// cached-projection claim: the fold is shared by every viewer, and the
// per-viewer difference is applied after it, from the board as it is at
// the time of the view. A card that becomes known to ONE seat between
// two views must flip that seat's line and no other's, with the fold
// untouched in between.
func TestCachedLogFollowsWhoKnowsTheCard(t *testing.T) {
	g := buildFourPlayerBoard(t)
	owner, friend, other := g.Seats[0], g.Seats[1], g.Seats[2]

	var cardID uuid.UUID
	var name string
	g.WithWriteLock(func() {
		hand := owner.Hand
		c := &hand.Cards[0]
		cardID, name = c.InstanceID, c.Name
		g.EmitEvent(game.Event{
			Kind: game.EventCast, Actor: owner.ID, Source: cardID, CardID: cardID,
			OldZone: game.ZoneHand, NewZone: game.ZoneStack,
		})
	})

	named := func(v GameView, viewer string) bool {
		for _, e := range FilterViewFor(v, viewer).Log {
			if e.CardID == cardID.String() && e.Kind == LogCast {
				return strings.Contains(e.Text, name)
			}
		}
		t.Fatalf("viewer %s: the cast is not in the log", viewer)
		return false
	}

	v := viewAndCheck(t, g, "cast of a card only its owner knows")
	var folded int
	g.LogProjectionCache().Do(func(slot *any) { folded = (*slot).(*logCache).folded })
	if named(v, other.ID.String()) || named(v, friend.ID.String()) {
		t.Fatalf("a seat that cannot see the card read its name %q in the log", name)
	}
	if !named(v, owner.ID.String()) {
		t.Fatalf("the owner should read the card's name in their own log")
	}

	g.WithWriteLock(func() {
		hand := owner.Hand
		hand.Cards[0].AddKnower(friend.ID)
	})
	v = viewAndCheck(t, g, "after the card became known to one other seat")
	if !named(v, friend.ID.String()) {
		t.Fatalf("the seat the card was shown to should now read its name")
	}
	if named(v, other.ID.String()) {
		t.Fatalf("a seat the card was NOT shown to read its name %q", name)
	}
	if named(v, spectatorID) {
		t.Fatalf("a spectator read the name of a card only two seats know")
	}
	var after int
	g.LogProjectionCache().Do(func(slot *any) { after = (*slot).(*logCache).folded })
	if after != folded {
		t.Fatalf("a change in who knows a card re-folded %d events; the fold is viewer-independent", after-folded)
	}
}

// TestRevealWindowFoldsOnlyNewEvents is the cost claim: after the first
// view, a view folds only the events emitted since the last one.
func TestRevealWindowFoldsOnlyNewEvents(t *testing.T) {
	g := buildFourPlayerBoard(t)
	growLongGame(t, g, 2000)
	v := ViewOfGame(g)

	var c *logCache
	g.LogProjectionCache().Do(func(slot *any) { c, _ = (*slot).(*logCache) })
	if c == nil || c.reveals == nil {
		t.Fatal("no reveal fold was kept")
	}
	first := c.reveals
	if first.cursor != len(g.Events) {
		t.Fatalf("fold is at %d, game has %d events", first.cursor, len(g.Events))
	}

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: g.Seats[0].ID})
	})
	_ = publicRevealsOf(g, &v)
	g.LogProjectionCache().Do(func(slot *any) { c, _ = (*slot).(*logCache) })
	if c.reveals != first {
		t.Fatal("a view after one new event rebuilt the fold instead of resuming it")
	}
	if c.reveals.cursor != len(g.Events) {
		t.Fatalf("fold is at %d, game has %d events", c.reveals.cursor, len(g.Events))
	}
}
