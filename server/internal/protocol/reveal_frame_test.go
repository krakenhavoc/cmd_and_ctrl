package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_frame_test.go is the executable half of the argument in
// reveal_frame.go.
//
// Two claims are being pinned, and they pull in opposite directions,
// which is why both need tests:
//
//  1. A reveal REACHES EVERY SEAT, including seats the ordinary zone
//     filter has stripped the source zone from entirely. An opponent's
//     library is wholesale-hidden; the reveal has to name cards out of
//     it anyway, or the frame does nothing.
//
//  2. A reveal HANDS OUT NO HANDLE. Not the revealed cards' instance
//     IDs, not the revealing card's. The same "a UUID is a correlation
//     handle" argument the public log settled and #513 had to settle
//     again for pending-choice options, arriving here third.

// revealTopOf reveals the top n cards of a seat's library and returns
// their instance IDs.
func revealTopOf(tb testing.TB, g *game.Game, seat *game.Player, source uuid.UUID, n int, reason string) []uuid.UUID {
	tb.Helper()
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		// Stamp the reveal onto the turn the view will report as
		// current, so the this-turn window keeps it.
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: seat.ID,
			Amount: g.Turn.Seq, Label: string(g.Turn.Step),
		})
		ids = g.RevealTopOfLibraryForEffect(seat.ID, source, n, reason)
	})
	if len(ids) != n {
		tb.Fatalf("revealed %d cards, want %d", len(ids), n)
	}
	return ids
}

func TestRevealFrameReachesEverySeatIdentically(t *testing.T) {
	g := buildFourPlayerBoard(t)
	revealer := g.Seats[0]
	revealTopOf(t, g, revealer, uuid.Nil, 3, "Test Source — reveal the top three")

	v := ViewOfGame(g)
	var want string
	for i, seat := range g.Seats {
		got := FilterViewFor(v, seat.ID.String())
		if len(got.Reveals) != 1 {
			t.Fatalf("seat %d sees %d reveals, want 1", i, len(got.Reveals))
		}
		if n := len(got.Reveals[0].Cards); n != 3 {
			t.Fatalf("seat %d sees %d revealed cards, want 3", i, n)
		}
		for _, c := range got.Reveals[0].Cards {
			if c.Name == "" {
				t.Errorf("seat %d got an unnamed card in a reveal — the frame was redacted", i)
			}
		}
		b, err := json.Marshal(got.Reveals)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if i == 0 {
			want = string(b)
			continue
		}
		if string(b) != want {
			t.Errorf("seat %d's reveal frame differs from seat 0's:\n got %s\nwant %s", i, b, want)
		}
	}

	// And a spectator, who has no seat at all.
	spectator := FilterViewFor(v, "")
	if len(spectator.Reveals) != 1 {
		t.Fatalf("a spectator sees %d reveals, want 1", len(spectator.Reveals))
	}
}

// TestRevealFrameNamesCardsInAZoneTheViewerCannotSee is claim 1. The
// ordinary filter leaves an opponent's library with zero cards on the
// wire; the reveal has to work anyway.
func TestRevealFrameNamesCardsInAZoneTheViewerCannotSee(t *testing.T) {
	g := buildFourPlayerBoard(t)
	revealer, watcher := g.Seats[0], g.Seats[1]
	revealTopOf(t, g, revealer, uuid.Nil, 2, "Test Source — reveal the top two")

	got := FilterViewFor(ViewOfGame(g), watcher.ID.String())
	for _, s := range got.Seats {
		if s.ID == revealer.ID.String() && len(s.Library.Cards) != 0 {
			t.Fatalf("the revealer's library is not hidden from the watcher (%d cards on the wire) — "+
				"this test is no longer testing what it thinks it is", len(s.Library.Cards))
		}
	}
	if len(got.Reveals) != 1 || len(got.Reveals[0].Cards) != 2 {
		t.Fatalf("the watcher did not receive the reveal: %+v", got.Reveals)
	}
	for _, c := range got.Reveals[0].Cards {
		if c.Name == "" {
			t.Error("a revealed card reached the watcher unnamed")
		}
	}
}

// TestRevealFrameCarriesNoInstanceIDs is claim 2, and it is the test
// this whole field was designed around. If it ever fails, the fix is
// to remove whatever field was added — not to relax the assertion.
func TestRevealFrameCarriesNoInstanceIDs(t *testing.T) {
	g := buildFourPlayerBoard(t)
	revealer := g.Seats[0]

	// A source card on the battlefield, so the frame has a real ID it
	// could have leaked rather than a nil one it could not.
	var sourceID uuid.UUID
	g.WithWriteLock(func() {
		sourceID = uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: sourceID,
			Name:       "Reveal Engine",
			TypeLine:   "Creature — Human Wizard",
			Owner:      revealer.ID,
			Controller: revealer.ID,
			KnownBy:    map[uuid.UUID]bool{revealer.ID: true},
		})
	})
	ids := revealTopOf(t, g, revealer, sourceID, 5, "Reveal Engine — reveal the top five")

	for _, seat := range g.Seats {
		got := FilterViewFor(ViewOfGame(g), seat.ID.String())
		b, err := json.Marshal(got.Reveals)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		blob := string(b)
		for _, id := range ids {
			if strings.Contains(blob, id.String()) {
				t.Fatalf("the reveal frame ships the instance ID of a revealed card (%s). "+
					"That is a correlation handle on a card sitting in a hidden zone — "+
					"the exact leak #513 closed for pending-choice options.", id)
			}
		}
		if strings.Contains(blob, sourceID.String()) {
			t.Errorf("the reveal frame ships the revealing card's instance ID (%s); its NAME is the public fact", sourceID)
		}
		// Sanity: the frame is not empty, so the assertions above are
		// not passing by saying nothing.
		if !strings.Contains(blob, "Reveal Engine") {
			t.Errorf("the frame does not name its source; the leak assertions above prove nothing:\n%s", blob)
		}
	}
}

// TestRevealFrameIsIdempotentUnderRepeatedFiltering mirrors the
// guarantee redactCardForViewer gives: filtering an already-filtered
// view must not change anything.
func TestRevealFrameIsIdempotentUnderRepeatedFiltering(t *testing.T) {
	g := buildFourPlayerBoard(t)
	revealTopOf(t, g, g.Seats[0], uuid.Nil, 3, "Test Source — reveal the top three")

	viewer := g.Seats[1].ID.String()
	once := FilterViewFor(ViewOfGame(g), viewer)
	twice := FilterViewFor(once, viewer)

	a, _ := json.Marshal(once.Reveals)
	b, _ := json.Marshal(twice.Reveals)
	if string(a) != string(b) {
		t.Errorf("a second filtering changed the reveal frame:\n once %s\ntwice %s", a, b)
	}
}

// TestRevealWindowIsBoundedAndThisTurn is the "for how long" test.
func TestRevealWindowIsBoundedAndThisTurn(t *testing.T) {
	g := buildFourPlayerBoard(t)
	revealer := g.Seats[0]

	// One reveal on a previous turn.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: revealer.ID,
			Amount: g.Turn.Seq - 1, Label: string(g.Turn.Step),
		})
		g.RevealTopOfLibraryForEffect(revealer.ID, uuid.Nil, 1, "last turn")
	})
	// Then PublicRevealMax+2 on this one.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: revealer.ID,
			Amount: g.Turn.Seq, Label: string(g.Turn.Step),
		})
		for i := 0; i < PublicRevealMax+2; i++ {
			g.RevealTopOfLibraryForEffect(revealer.ID, uuid.Nil, 1, fmt.Sprintf("this turn #%d", i))
		}
	})

	v := FilterViewFor(ViewOfGame(g), revealer.ID.String())
	if len(v.Reveals) != PublicRevealMax {
		t.Fatalf("window holds %d reveals, want the cap of %d", len(v.Reveals), PublicRevealMax)
	}
	for _, r := range v.Reveals {
		if r.Turn != g.Turn.Seq {
			t.Errorf("a reveal from turn %d survived into turn %d's window", r.Turn, g.Turn.Seq)
		}
		if r.Reason == "last turn" {
			t.Error("last turn's reveal is still on the wire")
		}
	}
	// Oldest first, and it is the OLDEST of the survivors that was
	// dropped — the window keeps the most recent.
	if got := v.Reveals[0].Reason; got != "this turn #2" {
		t.Errorf("window starts at %q, want the third of this turn's six reveals", got)
	}
	if got := v.Reveals[len(v.Reveals)-1].Reason; got != "this turn #5" {
		t.Errorf("window ends at %q, want the most recent reveal", got)
	}
	if v.Reveals[0].Seq >= v.Reveals[len(v.Reveals)-1].Seq {
		t.Error("the window is not in ascending sequence order")
	}
}

// TestRevealCardsAreCappedButCounted is the Hermit Druid case: reveal
// a whole library, ship a readable slice of it, and say how many
// there really were.
func TestRevealCardsAreCappedButCounted(t *testing.T) {
	g := buildFourPlayerBoard(t)
	revealer := g.Seats[0]
	total := revealer.Library.Size()
	if total <= RevealCardsMax {
		t.Fatalf("library is only %d cards; this test needs more than the %d-card cap", total, RevealCardsMax)
	}
	revealTopOf(t, g, revealer, uuid.Nil, total, "Test Source — reveal the whole library")

	v := FilterViewFor(ViewOfGame(g), g.Seats[1].ID.String())
	if len(v.Reveals) != 1 {
		t.Fatalf("got %d reveals, want 1", len(v.Reveals))
	}
	r := v.Reveals[0]
	if len(r.Cards) != RevealCardsMax {
		t.Errorf("shipped %d cards, want the cap of %d", len(r.Cards), RevealCardsMax)
	}
	if r.Count != total {
		t.Errorf("Count says %d, want the true total %d", r.Count, total)
	}
}

// TestRevealFrameSurvivesSnapshotRestore is why this field needed no
// entry in snapshot_drift_test.go: it is a projection of
// game.Game.Events, which that test already classifies `carried`. The
// window is derived state, so there is nothing new to carry — but
// Event grew a RevealSeq field to group the run, and this is the
// claim that the grouping survives a deploy.
func TestRevealFrameSurvivesSnapshotRestore(t *testing.T) {
	g := buildFourPlayerBoard(t)
	revealTopOf(t, g, g.Seats[0], uuid.Nil, 3, "Test Source — reveal the top three")
	before := ViewOfGame(g).Reveals

	snap := g.CaptureSnapshot()
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	after := ViewOfGame(restored).Reveals

	b, _ := json.Marshal(before)
	a, _ := json.Marshal(after)
	if string(a) != string(b) {
		t.Errorf("the reveal window drifted across a snapshot restore:\nbefore %s\n after %s", b, a)
	}
	if len(before) != 1 {
		t.Fatalf("the fixture produced %d reveals, want 1", len(before))
	}
}

// revealFrameBudget is what the reveal window may cost a four-player
// frame.
//
// The share, not a ratio: the public log is gated at 32 KiB and
// legal_moves at 24 KiB against a ~100 KiB four-player snapshot, and
// this field is a fraction of either. Its worst case is knowable in
// advance — PublicRevealMax reveals of RevealCardsMax cards each, so
// 48 entries — which is why the gate is an absolute and not a
// percentage.
//
// If this fails, the answer is NOT to raise the constant. It is to
// lower RevealCardsMax or PublicRevealMax, or to stop putting a field
// on RevealedCardView that the client does not draw.
const revealFrameBudget = 8 * 1024

func TestRevealFrameWireCost(t *testing.T) {
	g := buildFourPlayerBoard(t)
	revealer := g.Seats[0]
	// Real cards, not the board builder's bare fillers. A reveal
	// entry is almost entirely strings — name, mana cost, type line,
	// printing id — so measuring it against cards that have none of
	// those would flatter the field by a factor of four.
	g.WithWriteLock(func() {
		for i := range PublicRevealMax * RevealCardsMax {
			revealer.Library.PushTop(game.Card{
				InstanceID: uuid.New(),
				Name:       fmt.Sprintf("Thassa's Oracle of the Endless Tide %d", i),
				TypeLine:   "Legendary Enchantment Creature — Merfolk Wizard",
				ManaCost:   "{3}{W}{U}{B}{R}{G}",
				ScryfallID: uuid.NewString(),
				Owner:      revealer.ID,
				Controller: revealer.ID,
			})
		}
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: revealer.ID,
			Amount: g.Turn.Seq, Label: string(g.Turn.Step),
		})
		for range PublicRevealMax {
			g.RevealTopOfLibraryForEffect(revealer.ID, uuid.Nil, RevealCardsMax,
				fmt.Sprintf("Saturating Source — reveal the top %d cards of your library", RevealCardsMax))
		}
	})

	v := FilterViewFor(ViewOfGame(g), revealer.ID.String())
	reveals, entries := len(v.Reveals), 0
	for _, r := range v.Reveals {
		entries += len(r.Cards)
	}
	if reveals != PublicRevealMax || entries != PublicRevealMax*RevealCardsMax {
		t.Fatalf("window is not saturated: %d reveals / %d cards", reveals, entries)
	}

	withReveals, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	v.Reveals = nil
	without, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	delta := len(withReveals) - len(without)
	t.Logf("four-player frame: %d B without reveals, %d B with a saturated window (+%d B, +%.1f%%; %d B/card across %d cards in %d reveals)",
		len(without), len(withReveals), delta,
		100*float64(delta)/float64(len(without)), delta/entries, entries, reveals)
	t.Logf("deflated (what permessage-deflate would cost if the hub ever enables it): %d B -> %d B",
		deflatedLen(t, without), deflatedLen(t, withReveals))

	if delta > revealFrameBudget {
		t.Errorf("the reveal window costs %d B on the wire, budget is %d B — lower RevealCardsMax or PublicRevealMax rather than raising this",
			delta, revealFrameBudget)
	}
}

// BenchmarkPublicRevealsProjection is the per-FRAME half: the
// projection runs once per broadcast, over every event the game has
// emitted.
func BenchmarkPublicRevealsProjection(b *testing.B) {
	g := buildFourPlayerBoardB(b)
	revealer := g.Seats[0]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: revealer.ID,
			Amount: g.Turn.Seq, Label: string(g.Turn.Step),
		})
		for i := 0; i < PublicRevealMax; i++ {
			g.RevealTopOfLibraryForEffect(revealer.ID, uuid.Nil, RevealCardsMax, "bench")
		}
	})
	v := ViewOfGame(g)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = publicRevealsOf(g, &v)
	}
}
