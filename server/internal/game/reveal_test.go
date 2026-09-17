package game

import (
	"testing"

	"github.com/google/uuid"
)

// reveal_test.go pins the engine half of CR 701.20.
//
// The tests that matter here are the ones that would still pass if
// "reveal" had been built as "look at" with a bigger audience — which
// is the mistake the primitive exists to prevent. So:
//
//   - every seat becomes a knower, not just the revealer;
//   - nothing moves, because a reveal is not a zone change;
//   - the knowledge is STICKY, which is what "the table is entitled
//     to remember" means, and a shuffle is the thing that ends it;
//   - the announcement is one grouped run of events, so the wire can
//     say "Alice revealed these five cards" rather than five separate
//     things that happened to be adjacent.

// revealEventsFor returns the EventRevealCards entries stamped with
// the given RevealSeq, in emission order.
func revealEventsFor(g *Game, seq uint64) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventRevealCards && ev.RevealSeq == seq {
			out = append(out, ev)
		}
	}
	return out
}

// libraryCardByID finds a card in a player's library by instance ID.
func libraryCardByID(p *Player, id uuid.UUID) *Card {
	for i := range p.Library.Cards {
		if p.Library.Cards[i].InstanceID == id {
			return &p.Library.Cards[i]
		}
	}
	return nil
}

func TestRevealMarksEverySeatAKnower(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]

	g.mu.Lock()
	ids := g.RevealTopOfLibraryForEffect(me.ID, uuid.Nil, 1, "test — reveal the top card")
	g.mu.Unlock()

	if len(ids) != 1 {
		t.Fatalf("revealed %d cards, want 1", len(ids))
	}
	card := libraryCardByID(me, ids[0])
	if card == nil {
		t.Fatal("the revealed card left the library; a reveal does not move it")
	}
	for _, seat := range g.Seats {
		if !card.IsKnownTo(seat.ID) {
			t.Errorf("seat %s is not a knower of a card revealed to the table", seat.Name)
		}
	}
}

// TestRevealIsNotLookAt is the discriminator against the scry family.
// A scry marks the chooser and nobody else; this must mark everybody,
// and the two must not be the same code path by accident.
func TestRevealIsNotLookAt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]

	g.mu.Lock()
	g.ScryForEffect(me.ID, uuid.Nil, 1)
	scryed := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	g.mu.Unlock()

	looked := libraryCardByID(me, scryed)
	if looked == nil {
		t.Fatal("the scryed card left the library")
	}
	if !looked.IsKnownTo(me.ID) {
		t.Error("the scrying player is not a knower of their own top card")
	}
	for _, seat := range g.Seats[1:] {
		if looked.IsKnownTo(seat.ID) {
			t.Fatalf("seat %s knows a card that was only LOOKED AT — "+
				"reveal and look-at have collapsed into one path", seat.Name)
		}
	}
}

// TestRevealSurvivesTheCardReachingHand is the Dark Confidant shape:
// reveal the top card, then put it into your hand. Every seat saw it,
// so every seat keeps reading it in a zone they otherwise cannot.
func TestRevealSurvivesTheCardReachingHand(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	g.mu.Lock()
	ids := g.RevealTopOfLibraryForEffect(me.ID, uuid.Nil, 1, "test — reveal the top card")
	if len(ids) != 1 {
		g.mu.Unlock()
		t.Fatalf("revealed %d cards, want 1", len(ids))
	}
	if err := g.BounceToHandForEffect(ids[0]); err != nil {
		g.mu.Unlock()
		t.Fatalf("BounceToHandForEffect: %v", err)
	}
	g.mu.Unlock()

	var inHand *Card
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == ids[0] {
			inHand = &me.Hand.Cards[i]
		}
	}
	if inHand == nil {
		t.Fatal("the revealed card did not reach hand")
	}
	if !inHand.IsKnownTo(opp.ID) {
		t.Error("the opponent forgot a card they watched being revealed on its way to hand")
	}
}

// TestShuffleEndsWhatARevealGranted is the "for how long" test. A
// reveal is permanent knowledge right up until the library is
// randomised, and then it is not — which is the same rule scry lives
// under and the reason revealing a card that goes back on top of a
// library about to be shuffled is not a permanent information leak.
func TestShuffleEndsWhatARevealGranted(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	g.mu.Lock()
	ids := g.RevealTopOfLibraryForEffect(me.ID, uuid.Nil, 1, "test — reveal the top card")
	g.mu.Unlock()
	if len(ids) != 1 {
		t.Fatalf("revealed %d cards, want 1", len(ids))
	}
	if c := libraryCardByID(me, ids[0]); c == nil || !c.IsKnownTo(opp.ID) {
		t.Fatal("the reveal did not take")
	}

	if err := g.ShuffleLibrary(me.ID); err != nil {
		t.Fatalf("ShuffleLibrary: %v", err)
	}
	if c := libraryCardByID(me, ids[0]); c != nil && c.IsKnownTo(opp.ID) {
		t.Error("a shuffle did not end the knowledge a reveal granted")
	}
}

// TestRevealEmitsOneGroupedAnnouncement is what lets the wire say
// "revealed these five cards" instead of five unrelated things.
func TestRevealEmitsOneGroupedAnnouncement(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	g.mu.Lock()
	first := g.RevealForEffect(RevealSpec{
		Player: me.ID,
		Reason: "test — reveal the top five",
		Cards:  topOfLibrary(me, 5),
	})
	second := g.RevealForEffect(RevealSpec{
		Player: me.ID,
		Reason: "test — a second, separate reveal",
		Cards:  topOfLibrary(me, 1),
	})
	g.mu.Unlock()

	if first == 0 || second == 0 {
		t.Fatalf("reveal returned no group key (first=%d second=%d)", first, second)
	}
	if first == second {
		t.Fatal("two separate reveals share a group key — the wire will merge them into one announcement")
	}
	if n := len(revealEventsFor(g, first)); n != 5 {
		t.Errorf("first reveal emitted %d events, want 5 (one per card)", n)
	}
	if n := len(revealEventsFor(g, second)); n != 1 {
		t.Errorf("second reveal emitted %d events, want 1", n)
	}
	for _, ev := range revealEventsFor(g, first) {
		if ev.OldZone != ZoneLibrary {
			t.Errorf("reveal event carries OldZone %q, want the zone it was revealed from", ev.OldZone)
		}
		if ev.NewZone != "" {
			t.Errorf("reveal event carries NewZone %q — a reveal is not a zone change", ev.NewZone)
		}
	}
}

// TestRevealSeqIsDeterministic guards the choice of an event sequence
// number over a freshly minted UUID: a replayed game has to produce
// the same event stream byte for byte.
func TestRevealSeqIsDeterministic(t *testing.T) {
	seqOf := func() uint64 {
		g := newActiveGame(t)
		me := g.Seats[0]
		g.mu.Lock()
		defer g.mu.Unlock()
		return g.RevealForEffect(RevealSpec{Player: me.ID, Cards: topOfLibrary(me, 2)})
	}
	if a, b := seqOf(), seqOf(); a != b {
		t.Errorf("the same game produced RevealSeq %d then %d; a replay would diverge", a, b)
	}
}

func TestRevealTopOfEmptyOrShortLibrary(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	size := me.Library.Size()
	g.mu.Lock()
	over := g.RevealTopOfLibraryForEffect(me.ID, uuid.Nil, size+10, "test — reveal more than there is")
	g.mu.Unlock()
	if len(over) != size {
		t.Errorf("revealed %d of a %d-card library, want all of it", len(over), size)
	}

	me.Library.Cards = nil
	g.mu.Lock()
	none := g.RevealTopOfLibraryForEffect(me.ID, uuid.Nil, 1, "test — empty library")
	g.mu.Unlock()
	if len(none) != 0 {
		t.Errorf("revealed %d cards off an empty library", len(none))
	}
	if me.AttemptedEmptyDraw {
		t.Error("revealing off an empty library set up the draw-from-empty loss; revealing is not drawing")
	}
}

// TestRevealSkipsCardsThatHaveLeft — a reveal is a look, and a look
// at something that is no longer there is nothing, not an error.
func TestRevealSkipsCardsThatHaveLeft(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	real := topOfLibrary(me, 1)
	g.mu.Lock()
	seq := g.RevealForEffect(RevealSpec{
		Player: me.ID,
		Cards:  []uuid.UUID{uuid.New(), real[0], uuid.New()},
	})
	g.mu.Unlock()
	if n := len(revealEventsFor(g, seq)); n != 1 {
		t.Errorf("emitted %d events, want 1 — the two phantom IDs should be skipped", n)
	}
}

// topOfLibrary returns the top n instance IDs of a player's library,
// top card first.
func topOfLibrary(p *Player, n int) []uuid.UUID {
	size := p.Library.Size()
	if n > size {
		n = size
	}
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, p.Library.Cards[size-1-i].InstanceID)
	}
	return out
}
