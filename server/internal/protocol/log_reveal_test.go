package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_reveal_test.go pins the LogReveal entry: a reveal leaves one
// line in the public log however many cards it showed, the line names
// what the table saw, and no viewer ever gets an instance ID for a
// card revealed out of a hidden zone.

// revealLogEntries returns every LogReveal entry in a log.
func revealLogEntries(entries []LogEvent) []LogEvent {
	var out []LogEvent
	for _, e := range entries {
		if e.Kind == LogReveal {
			out = append(out, e)
		}
	}
	return out
}

// allViewers is every viewer ID a log is filtered for: the empty
// (spectator / admin / replay) viewer and each seat.
func allViewers(g *game.Game) []string {
	out := []string{""}
	for _, p := range g.Seats {
		out = append(out, p.ID.String())
	}
	return out
}

// topNames returns the names of the top n library cards, top first.
func topNames(g *game.Game, p *game.Player, n int) []string {
	var out []string
	g.ReadSnapshot(func() {
		size := p.Library.Size()
		for i := 0; i < n; i++ {
			out = append(out, p.Library.Cards[size-1-i].Name)
		}
	})
	return out
}

// nextEventSeq returns the Seq the next EmitEvent will stamp, which is
// the RevealSeq game.RevealForEffect would key a reveal on. It anchors
// the log on a step announcement first, so there is a last event to
// read. Caller must hold the write lock.
func nextEventSeq(g *game.Game) uint64 {
	g.EmitEvent(game.Event{Kind: game.EventStepBegan, Amount: 1, Label: string(game.StepUpkeep)})
	return g.Events[len(g.Events)-1].Seq + 1
}

// assertNoInstanceIDs fails when any of ids appears anywhere in the
// marshalled log.
func assertNoInstanceIDs(t *testing.T, viewer string, log []LogEvent, ids []uuid.UUID) {
	t.Helper()
	buf, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("marshal log: %v", err)
	}
	for _, id := range ids {
		if strings.Contains(string(buf), id.String()) {
			t.Errorf("viewer %q: revealed card's instance ID %s is on the log wire: %s", viewer, id, buf)
		}
	}
}

func TestPublicLogRevealIsOneLineWithNoCardID(t *testing.T) {
	g := buildActiveGame(t)
	revealer := g.Seats[0]
	want := topNames(g, revealer, 3)

	var ids []uuid.UUID
	g.WithWriteLock(func() {
		ids = g.RevealTopOfLibraryForEffect(revealer.ID, uuid.Nil, 3, "Fact or Fiction")
	})
	if len(ids) != 3 {
		t.Fatalf("revealed %d cards, want 3", len(ids))
	}
	// Then the cards go back to being hidden: a shuffle clears every
	// seat's knowledge of them. The history must survive that.
	if err := g.ShuffleLibrary(revealer.ID); err != nil {
		t.Fatalf("ShuffleLibrary: %v", err)
	}

	wantText := fmt.Sprintf("P1 revealed 3 cards from their library: %s", strings.Join(want, ", "))
	for _, viewer := range allViewers(g) {
		log := ViewOfGameFor(g, viewer).Log
		reveals := revealLogEntries(log)
		if len(reveals) != 1 {
			t.Fatalf("viewer %q: %d reveal entries, want 1 collapsed entry: %+v", viewer, len(reveals), reveals)
		}
		e := reveals[0]
		if e.CardID != "" || e.Target != "" {
			t.Errorf("viewer %q: reveal entry carries a card reference: %+v", viewer, e)
		}
		if e.Amount != 3 || e.Seat != revealer.Seat || e.OldZone != string(game.ZoneLibrary) || e.TargetSeat != nil {
			t.Errorf("viewer %q: reveal entry fields: %+v", viewer, e)
		}
		if e.Text != wantText {
			t.Errorf("viewer %q: text %q, want %q", viewer, e.Text, wantText)
		}
		assertNoInstanceIDs(t, viewer, log, ids)
	}
}

// TestPublicLogRevealDarkConfidantLogsOnce is the common shape: one
// card revealed, then moved library → hand. One reveal line, no zone
// line, and no handle to follow the card into the hand with.
func TestPublicLogRevealDarkConfidantLogsOnce(t *testing.T) {
	g := buildActiveGame(t)
	p := g.Seats[0]
	name := topNames(g, p, 1)[0]
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		ids = g.RevealTopOfLibraryForEffect(p.ID, uuid.Nil, 1, "Dark Confidant")
	})
	err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneLibrary, Owner: p.ID},
		game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID},
		ids[0],
	)
	if err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	for _, viewer := range allViewers(g) {
		log := ViewOfGameFor(g, viewer).Log
		reveals := revealLogEntries(log)
		if len(reveals) != 1 {
			t.Fatalf("viewer %q: %d reveal entries, want 1", viewer, len(reveals))
		}
		if want := fmt.Sprintf("P1 revealed %s from their library", name); reveals[0].Text != want {
			t.Errorf("viewer %q: text %q, want %q", viewer, reveals[0].Text, want)
		}
		for _, e := range log {
			if e.Kind == LogZone {
				t.Errorf("viewer %q: library → hand produced a zone entry: %+v", viewer, e)
			}
		}
		assertNoInstanceIDs(t, viewer, log, ids)
	}
}

// TestPublicLogRevealsCollapseByRevealSeqNotAdjacency: two reveals
// back to back are two lines, and one reveal with an unrelated event
// emitted between its cards is still one.
func TestPublicLogRevealsCollapseByRevealSeqNotAdjacency(t *testing.T) {
	g := buildActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.RevealTopOfLibraryForEffect(a.ID, uuid.Nil, 1, "each player reveals")
		g.RevealTopOfLibraryForEffect(b.ID, uuid.Nil, 1, "each player reveals")
	})
	reveals := revealLogEntries(ViewOfGame(g).Log)
	if len(reveals) != 2 {
		t.Fatalf("%d reveal entries for two adjacent reveals, want 2: %+v", len(reveals), reveals)
	}
	if reveals[0].Seat != a.Seat || reveals[1].Seat != b.Seat || reveals[0].Amount != 1 || reveals[1].Amount != 1 {
		t.Errorf("adjacent reveals merged or misattributed: %+v", reveals)
	}

	g2 := buildActiveGame(t)
	p := g2.Seats[0]
	var top []uuid.UUID
	g2.ReadSnapshot(func() {
		size := p.Library.Size()
		top = []uuid.UUID{p.Library.Cards[size-1].InstanceID, p.Library.Cards[size-2].InstanceID}
	})
	g2.WithWriteLock(func() {
		first := nextEventSeq(g2)
		g2.EmitEvent(game.Event{
			Kind: game.EventRevealCards, Actor: p.ID, CardID: top[0],
			OldZone: game.ZoneLibrary, RevealSeq: first,
		})
		g2.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: p.ID, Amount: 1})
		g2.EmitEvent(game.Event{
			Kind: game.EventRevealCards, Actor: p.ID, CardID: top[1],
			OldZone: game.ZoneLibrary, RevealSeq: first,
		})
	})
	log := ViewOfGame(g2).Log
	reveals = revealLogEntries(log)
	if len(reveals) != 1 || reveals[0].Amount != 2 {
		t.Fatalf("interleaved reveal did not collapse to one 2-card entry: %+v", reveals)
	}
	findLog(t, log, LogLife)
}

// TestPublicLogPrivateRevealNamesNothing covers a reveal with a player
// target, which only that player saw. Every viewer gets the line, and
// no viewer gets the card's name or instance ID from it.
func TestPublicLogPrivateRevealNamesNothing(t *testing.T) {
	g := buildActiveGame(t)
	shower, shown := g.Seats[0], g.Seats[1]
	secret := uuid.New()
	g.WithWriteLock(func() {
		shower.Hand.PushTop(game.Card{
			InstanceID: secret,
			Name:       "Secret Card Name",
			Owner:      shower.ID,
			Controller: shower.ID,
			KnownBy:    map[uuid.UUID]bool{shower.ID: true, shown.ID: true},
		})
		seq := nextEventSeq(g)
		g.EmitEvent(game.Event{
			Kind: game.EventRevealCards, Actor: shower.ID, Target: shown.ID, CardID: secret,
			OldZone: game.ZoneHand, RevealSeq: seq,
		})
	})
	for _, viewer := range allViewers(g) {
		log := ViewOfGameFor(g, viewer).Log
		reveals := revealLogEntries(log)
		if len(reveals) != 1 {
			t.Fatalf("viewer %q: %d reveal entries, want 1", viewer, len(reveals))
		}
		e := reveals[0]
		if e.TargetSeat == nil || *e.TargetSeat != shown.Seat {
			t.Errorf("viewer %q: target seat %v, want %d", viewer, e.TargetSeat, shown.Seat)
		}
		if e.CardID != "" {
			t.Errorf("viewer %q: private reveal carries a card_id: %+v", viewer, e)
		}
		if want := "P1 revealed a card from their hand to P2"; e.Text != want {
			t.Errorf("viewer %q: text %q, want %q", viewer, e.Text, want)
		}
		buf, err := json.Marshal(log)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if strings.Contains(string(buf), "Secret Card Name") {
			t.Errorf("viewer %q: private reveal named the card on the log wire: %s", viewer, buf)
		}
		assertNoInstanceIDs(t, viewer, log, []uuid.UUID{secret})
	}
}

// TestPublicLogRevealTextIsCapped is the Hermit Druid case: the count
// is exact, the names stop at logRevealNamesMax.
func TestPublicLogRevealTextIsCapped(t *testing.T) {
	// The two-player fixture's libraries are three cards deep after the
	// opening draw; the four-player board's are not.
	g := buildFourPlayerBoard(t)
	p := g.Seats[0]
	const n = logRevealNamesMax + 3
	names := topNames(g, p, logRevealNamesMax)
	g.WithWriteLock(func() {
		g.RevealTopOfLibraryForEffect(p.ID, uuid.Nil, n, "Hermit Druid")
	})
	reveals := revealLogEntries(ViewOfGame(g).Log)
	if len(reveals) != 1 || reveals[0].Amount != n {
		t.Fatalf("want one %d-card entry, got %+v", n, reveals)
	}
	want := fmt.Sprintf("Player 1 revealed %d cards from their library: %s and 3 more", n, strings.Join(names, ", "))
	if reveals[0].Text != want {
		t.Errorf("text %q, want %q", reveals[0].Text, want)
	}
}

// TestLogRingPushedTracksEviction pins the ordinal lookup the reveal
// collapse leans on.
func TestLogRingPushedTracksEviction(t *testing.T) {
	r := newLogRing(3)
	for i := range 5 {
		r.push(LogEvent{Seq: uint64(i)})
	}
	for ord := range 5 {
		got := r.pushed(ord)
		if ord < 2 {
			if got != nil {
				t.Errorf("ordinal %d was evicted but pushed returned seq %d", ord, got.Seq)
			}
			continue
		}
		if got == nil || got.Seq != uint64(ord) {
			t.Errorf("ordinal %d: got %+v, want seq %d", ord, got, ord)
		}
	}
	if r.pushed(5) != nil {
		t.Error("an ordinal never pushed returned an entry")
	}
}
