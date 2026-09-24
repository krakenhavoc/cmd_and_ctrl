package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// library_order_test.go — ADR 0088's ordered library placement (#996):
// the put_in_library prompt, its three placements, the move behind it
// and who knows what afterwards.

// loSeed pushes named cards onto the TOP of `z` so the first name ends
// on top, and returns their IDs in the order given.
func loSeed(z *Zone, owner uuid.UUID, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i := range names {
		ids[i] = uuid.New()
	}
	for i := len(names) - 1; i >= 0; i-- {
		z.PushTop(Card{InstanceID: ids[i], Name: names[i], TypeLine: "Sorcery", Owner: owner, Controller: owner})
	}
	return ids
}

// loTop is the top n cards of a library, top-first.
func loTop(z *Zone, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, z.Cards[len(z.Cards)-1-i].InstanceID)
	}
	return out
}

// loBottom is the bottom n cards of a library, top-first (the last
// entry is the bottom card).
func loBottom(z *Zone, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := n - 1; i >= 0; i-- {
		out = append(out, z.Cards[i].InstanceID)
	}
	return out
}

func loPrompt(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == PendingChoicePutInLibrary {
			return c
		}
	}
	t.Fatal("no put_in_library prompt")
	return nil
}

func loQueue(t *testing.T, g *Game, spec PutInLibrarySpec) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.PutInLibraryInChosenOrderThenForEffect(spec); err != nil {
			t.Fatalf("PutInLibraryInChosenOrderThenForEffect: %v", err)
		}
	})
}

func loEqual(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// The bottom lane is top-first: the first card answered is the one
// nearest the top of the pile that goes under, the last is the bottom
// card. Cards already in the library are REORDERED — no zone move.
func TestPutInLibraryBottomIsTopFirstAndAReorder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := loSeed(me.Library, me.ID, "A", "B", "C")
	size := me.Library.Size()
	moves := countEventsOfKind(g, EventZoneMove)
	ran := 0
	loQueue(t, g, PutInLibrarySpec{
		Chooser: me.ID, Cards: ids, From: ZoneLibrary, Placement: LibraryPlaceBottom,
		Then: func(*Game) error { ran++; return nil },
	})
	c := loPrompt(t, g)
	if c.LibraryPlacement != LibraryPlaceBottom || len(c.ScryCards) != 3 {
		t.Fatalf("prompt %+v", c)
	}
	if ran != 0 {
		t.Fatal("Then must wait for the answer")
	}
	want := []uuid.UUID{ids[2], ids[0], ids[1]}
	if err := g.ResolvePutInLibrary(c.ID, me.ID, want, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := loBottom(me.Library, 3); !loEqual(got, want) {
		t.Errorf("bottom three %v, want %v", got, want)
	}
	if me.Library.Size() != size {
		t.Error("a reorder adds and removes nothing")
	}
	if n := countEventsOfKind(g, EventZoneMove); n != moves {
		t.Errorf("a reorder is not a zone change: %d zone moves emitted", n-moves)
	}
	if ran != 1 {
		t.Errorf("Then ran %d times, want once", ran)
	}
}

// Cards from a HAND go through the tuck route, top lane last-first so
// the answer's first card ends on top.
func TestPutInLibraryTopFromHandLandsInAnswerOrder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := loSeed(me.Hand, me.ID, "X", "Y", "Z")
	for i := range me.Hand.Cards {
		me.Hand.Cards[i].AddKnower(me.ID)
	}
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: ids, From: ZoneHand, Placement: LibraryPlaceTop})
	c := loPrompt(t, g)
	want := []uuid.UUID{ids[1], ids[2], ids[0]}
	if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := loTop(me.Library, 3); !loEqual(got, want) {
		t.Errorf("top three %v, want %v", got, want)
	}
	for _, id := range ids {
		if me.Hand.Contains(id) {
			t.Error("the cards left the hand")
		}
		card := g.findCardByIDLocked(id)
		if card == nil || !card.IsKnownTo(me.ID) {
			t.Error("ADR 0088 Decision 3: the chooser still knows what they put back")
		}
	}
}

// top_or_bottom partitions; each lane is ordered; a card put in no lane,
// two lanes, or a lane the placement does not open is refused.
func TestPutInLibraryAnswerShapeIsEnforced(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := loSeed(me.Library, me.ID, "A", "B", "C")
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: ids, From: ZoneLibrary, Placement: LibraryPlaceTopOrBottom})
	c := loPrompt(t, g)
	for name, ans := range map[string][2][]uuid.UUID{
		"missing a card":   {{ids[0]}, {ids[1]}},
		"a card twice":     {{ids[0], ids[1]}, {ids[1], ids[2]}},
		"a stranger":       {{ids[0], ids[1]}, {ids[2], uuid.New()}},
		"nothing answered": {nil, nil},
	} {
		if err := g.ResolvePutInLibrary(c.ID, me.ID, ans[0], ans[1]); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: err %v, want ErrInvalidParam", name, err)
		}
	}
	if err := g.ResolvePutInLibrary(c.ID, g.Seats[1].ID, nil, ids); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("another seat answering: %v", err)
	}
	if err := g.ResolvePutInLibrary(c.ID, me.ID, []uuid.UUID{ids[2], ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != ids[1] {
		t.Error("B kept on top")
	}
	if got := loBottom(me.Library, 2); !loEqual(got, []uuid.UUID{ids[2], ids[0]}) {
		t.Errorf("bottom two %v, want [C A]", got)
	}

	// A top placement refuses a bottom lane.
	more := loSeed(me.Hand, me.ID, "P", "Q")
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: more, From: ZoneHand, Placement: LibraryPlaceTop})
	c = loPrompt(t, g)
	if err := g.ResolvePutInLibrary(c.ID, me.ID, []uuid.UUID{more[0]}, []uuid.UUID{more[1]}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a top placement accepted a bottom lane: %v", err)
	}
}

// No prompt when there is no choice: an empty pile runs Then at once,
// and a pile of one with a one-lane placement is placed. top_or_bottom
// asks even about one card.
func TestPutInLibraryAsksOnlyWhenThereIsAChoice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ran := 0
	then := func(*Game) error { ran++; return nil }
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: nil, Placement: LibraryPlaceBottom, Then: then})
	if ran != 1 || len(g.PendingChoices) != 0 {
		t.Fatal("an empty pile runs Then and asks nothing")
	}
	one := loSeed(me.Hand, me.ID, "Solo")[0]
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: []uuid.UUID{one}, From: ZoneHand, Placement: LibraryPlaceBottom, Then: then})
	if len(g.PendingChoices) != 0 || ran != 2 {
		t.Fatal("a pile of one has no order")
	}
	if me.Library.Cards[0].InstanceID != one {
		t.Error("placed on the bottom")
	}
	again := loSeed(me.Hand, me.ID, "Hinder'd")[0]
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: []uuid.UUID{again}, From: ZoneHand, Placement: LibraryPlaceTopOrBottom})
	if c := loPrompt(t, g); len(c.ScryCards) != 1 {
		t.Error("top or bottom is a choice even for one card")
	}
}

// CR 400.7: a card that left the From zone between the prompt and the
// answer is skipped — not pulled back out of wherever it went.
func TestPutInLibrarySkipsACardThatLeftItsZone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := loSeed(me.Hand, me.ID, "Stays", "Leaves")
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: ids, From: ZoneHand, Placement: LibraryPlaceTop})
	c := loPrompt(t, g)
	moved, _ := me.Hand.Remove(ids[1])
	me.Graveyard.PushTop(moved)
	if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, []uuid.UUID{ids[1], ids[0]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if !me.Graveyard.Contains(ids[1]) {
		t.Error("the card that left stays where it went")
	}
	if top, _ := me.Library.Top(); top.InstanceID != ids[0] {
		t.Error("the one still in hand is placed")
	}
}

// ADR 0088 Decision 3 / CR 401.4: a card alone in its lane keeps its
// knowers; two or more to a lane are known by the chooser alone.
func TestPutInLibraryKnowersFollowCR4014(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := loSeed(me.Library, me.ID, "A", "B", "C")
	for i := range me.Library.Cards {
		me.Library.Cards[i].AddKnowersAll([]uuid.UUID{me.ID, opp.ID})
	}
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: ids, From: ZoneLibrary, Placement: LibraryPlaceTopOrBottom})
	c := loPrompt(t, g)
	if err := g.ResolvePutInLibrary(c.ID, me.ID, []uuid.UUID{ids[1], ids[2]}, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	alone := g.findCardByIDLocked(ids[0])
	if !alone.IsKnownTo(opp.ID) || !alone.IsKnownTo(me.ID) {
		t.Error("a card alone in its lane has no order to hide: it keeps its knowers")
	}
	for _, id := range ids[1:] {
		pile := g.findCardByIDLocked(id)
		if pile.IsKnownTo(opp.ID) {
			t.Error("CR 401.4: the order of two or more is not revealed")
		}
		if !pile.IsKnownTo(me.ID) {
			t.Error("the chooser knows the order they chose")
		}
	}
}

// A commander leg pauses on CR 903.9 and the pile waits for it: the
// card after it is placed from its continuation, so the order holds
// whichever way the owner answers.
func TestPutInLibraryCommanderLegKeepsThePileOrder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	cmd := seatCommander(t, g.Battlefield, me)
	plain := loSeed(g.Battlefield, me.ID, "Plain")[0]
	ran := 0
	loQueue(t, g, PutInLibrarySpec{
		Chooser: me.ID, Cards: []uuid.UUID{plain, cmd}, From: ZoneBattlefield, Placement: LibraryPlaceTop,
		Then: func(*Game) error { ran++; return nil },
	})
	c := loPrompt(t, g)
	// Answer [plain, cmd]: the top lane runs last-first, so cmd's leg
	// runs FIRST and pauses, and plain must wait for it.
	if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, []uuid.UUID{plain, cmd}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if !g.Battlefield.Contains(plain) {
		t.Fatal("the pile waits for the paused commander leg")
	}
	if ran != 0 {
		t.Fatal("Then waits for the whole pile")
	}
	prompt := expectCommanderPrompt(t, g, me)
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline the command zone: %v", err)
	}
	if got := loTop(me.Library, 2); !loEqual(got, []uuid.UUID{plain, cmd}) {
		t.Errorf("top two %v, want [plain commander]", got)
	}
	if ran != 1 {
		t.Errorf("Then ran %d times", ran)
	}
}

// Undo across the prompt: the continuation replays against the
// restored game and lands identically.
func TestPutInLibraryReplaysAfterUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := loSeed(me.Hand, me.ID, "X", "Y")
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: ids, From: ZoneHand, Placement: LibraryPlaceTop})
	open := g.Clone()
	answer := []uuid.UUID{ids[1], ids[0]}
	c := loPrompt(t, g)
	if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, answer); err != nil {
		t.Fatalf("first answer: %v", err)
	}
	first := loTop(g.Seats[0].Library, 2)
	g.WithWriteLock(func() { g.RestoreFrom(open) })
	me = g.Seats[0]
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[1]) {
		t.Fatal("the rewind puts the cards back in hand")
	}
	c = loPrompt(t, g)
	if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, answer); err != nil {
		t.Fatalf("replayed answer: %v", err)
	}
	if got := loTop(me.Library, 2); !loEqual(got, first) || !loEqual(got, answer) {
		t.Errorf("replay landed %v, first run %v, want %v", got, first, answer)
	}
}

// The snapshot carries the placement and counts the continuation.
func TestPutInLibraryIsNotARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := loSeed(me.Hand, me.ID, "X", "Y")
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: ids, From: ZoneHand, Placement: LibraryPlaceTop})
	snap := g.CaptureSnapshot()
	if snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 1 {
		t.Errorf("an open put_in_library is not a restore point: %#v", snap.Continuations)
	}
	found := false
	for _, c := range snap.PendingChoices {
		if c.Kind == PendingChoicePutInLibrary {
			found = true
			if c.LibraryPlacement != LibraryPlaceTop {
				t.Errorf("placement %q not carried", c.LibraryPlacement)
			}
			hasFrame := false
			for _, f := range c.ResumeFrames {
				if f == "libraryOrderResume" {
					hasFrame = true
				}
			}
			if !hasFrame {
				t.Error("the continuation is not counted in the census")
			}
		}
	}
	if !found {
		t.Fatal("the prompt is not in the snapshot")
	}
}
