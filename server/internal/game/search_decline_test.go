package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// search_decline_test.go — #4134, CR 701.19: "you may search" has an
// answer that is not "find nothing". A search that finds nothing is
// still a search (the library is shuffled, the search event fires); a
// declined search never happened (neither does).

func declineLibraryIDs(p *Player) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

func sameDeclineIDs(a, b []uuid.UUID) bool {
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

func optionalSwampSearch(me *Player, then func(*Game, []uuid.UUID) error) SearchLibrarySpec {
	return SearchLibrarySpec{
		Player:   me.ID,
		Pred:     func(c Card) bool { return c.Name == "Swamp" },
		Dest:     ZoneHand,
		Limit:    1,
		Shuffle:  true,
		Optional: true,
		ThenTail: searchHook(then),
	}
}

func TestDeclinedOptionalSearchDoesNotShuffleOrAnnounce(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	// Enough cards that a shuffle leaving the order alone is not a
	// plausible accident.
	names := []string{"Swamp"}
	for i := 0; i < 30; i++ {
		names = append(names, "Island")
	}
	ids := searchLibraryOf(me, names...)
	before := declineLibraryIDs(me)
	thenRan, thenFound := 0, -1

	runSearch(t, g, optionalSwampSearch(me, func(_ *Game, found []uuid.UUID) error {
		thenRan++
		thenFound = len(found)
		return nil
	}))
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt for an optional search")
	}
	if !c.SearchIsOptional() || !c.SearchIsUndecided() {
		t.Fatalf("optional/undecided = %v/%v, want true/true: an optional search asks before it looks", c.SearchIsOptional(), c.SearchIsUndecided())
	}
	if libraryKnownTo(me, me.ID) != 0 {
		t.Error("the searcher knows library cards before deciding to search: a free look")
	}

	if err := g.DeclineSearchLibrary(c.ID, me.ID); err != nil {
		t.Fatalf("declining a 'you may search': %v", err)
	}
	if openSearch(g, me.ID) != nil {
		t.Error("the prompt is still open after declining")
	}
	if !sameDeclineIDs(declineLibraryIDs(me), before) {
		t.Error("the library changed order: a declined search must not shuffle")
	}
	if me.Hand.Contains(ids[0]) {
		t.Error("declining still took the card")
	}
	if n := searchEvents(g); n != 0 {
		t.Errorf("EventSearchLibrary fired %d times for a search that never happened", n)
	}
	if thenRan != 1 || thenFound != 0 {
		t.Errorf("Then ran %d times with %d found, want once with none (the rest of the effect stands)", thenRan, thenFound)
	}
}

// The contrast the issue is about: accepting and finding nothing is a
// search, and still shuffles and announces.
func TestOptionalSearchThatFindsNothingStillShufflesAndAnnounces(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := searchLibraryOf(me, "Swamp", "Island", "Island")
	runSearch(t, g, optionalSwampSearch(me, nil))
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.BeginSearchLibrary(c.ID, me.ID); err != nil {
		t.Fatalf("beginning the search: %v", err)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("finding nothing: %v", err)
	}
	if me.Hand.Contains(ids[0]) {
		t.Error("the card was taken")
	}
	if n := searchEvents(g); n != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1: finding nothing is still a search", n)
	}
	if libraryKnownTo(me, me.ID) != 0 {
		t.Error("the library was not shuffled after finding nothing")
	}
}

// A mandatory search has no decline, and a refused decline leaves the
// prompt answerable.
func TestAMandatorySearchCannotBeDeclined(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Swamp", "Island")
	spec := optionalSwampSearch(me, nil)
	spec.Optional = false
	runSearch(t, g, spec)
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if c.SearchIsOptional() {
		t.Error("a mandatory search reports itself optional")
	}
	if err := g.DeclineSearchLibrary(c.ID, me.ID); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("declining a mandatory search = %v, want ErrInvalidParam", err)
	}
	if openSearch(g, me.ID) == nil {
		t.Error("the refused decline dropped the prompt")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Errorf("the prompt should still take 'find nothing': %v", err)
	}
}

func TestOnlyTheChooserMayDecline(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Swamp", "Island")
	runSearch(t, g, optionalSwampSearch(me, nil))
	c := openSearch(g, me.ID)
	if err := g.DeclineSearchLibrary(c.ID, g.Seats[1].ID); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("another seat declining = %v, want ErrNotTheChooser", err)
	}
	if err := g.BeginSearchLibrary(c.ID, g.Seats[1].ID); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("another seat beginning = %v, want ErrNotTheChooser", err)
	}
}

// Looking through the library IS searching (CR 701.19): once the searcher
// has said yes and been shown the candidates there is no way back. A
// decline at that point would be a free look that never shuffles.
func TestABegunSearchCannotBeDeclined(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Swamp", "Island")
	runSearch(t, g, optionalSwampSearch(me, nil))
	c := openSearch(g, me.ID)
	if c == nil || !c.SearchIsUndecided() {
		t.Fatalf("expected an undecided prompt: %+v", c)
	}
	// Undecided: no card pick is accepted, nothing is known.
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a pick before the search began = %v, want ErrInvalidParam", err)
	}
	if err := g.BeginSearchLibrary(c.ID, me.ID); err != nil {
		t.Fatalf("beginning: %v", err)
	}
	if libraryKnownTo(me, me.ID) == 0 {
		t.Error("the searcher should see the candidates once the search began")
	}
	if c.SearchIsUndecided() || g.SearchDeclineLegalLocked(c) {
		t.Error("a begun search still offers a way out")
	}
	if err := g.DeclineSearchLibrary(c.ID, me.ID); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("declining a begun search = %v, want ErrInvalidParam", err)
	}
	if err := g.BeginSearchLibrary(c.ID, me.ID); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("beginning twice = %v, want ErrInvalidParam", err)
	}
	if openSearch(g, me.ID) == nil {
		t.Fatal("a refused answer dropped the prompt")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("finding nothing: %v", err)
	}
	if n := searchEvents(g); n != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1: a search that looked has to shuffle", n)
	}
	if libraryKnownTo(me, me.ID) != 0 {
		t.Error("the library was not shuffled after the searcher looked")
	}
}

// --- #4140: nothing matches ---
//
// A search over a library with no match asks all the same: it used to
// resolve silently, which gave an optional search no way to be declined
// and let the table tell a library with a match from one without by the
// absence of a prompt.

func TestZeroMatchOptionalSearchCanBeDeclined(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	names := []string{}
	for i := 0; i < 30; i++ {
		names = append(names, "Island")
	}
	searchLibraryOf(me, names...)
	before := declineLibraryIDs(me)
	thenRan := 0

	runSearch(t, g, optionalSwampSearch(me, func(_ *Game, found []uuid.UUID) error {
		thenRan++
		return nil
	}))
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("a \"you may search\" with nothing to find did not ask")
	}
	if len(c.SearchCards) != 0 || !c.SearchIsOptional() {
		t.Fatalf("candidates=%d optional=%v, want none and true", len(c.SearchCards), c.SearchIsOptional())
	}
	if err := g.DeclineSearchLibrary(c.ID, me.ID); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if !sameDeclineIDs(declineLibraryIDs(me), before) {
		t.Error("the library changed order: a declined search must not shuffle")
	}
	if n := searchEvents(g); n != 0 {
		t.Errorf("EventSearchLibrary fired %d times for a search that never happened", n)
	}
	if thenRan != 1 {
		t.Errorf("Then ran %d times, want once", thenRan)
	}
}

func TestZeroMatchOptionalSearchThatFindsNothingStillShuffles(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Island", "Island")
	runSearch(t, g, optionalSwampSearch(me, nil))
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.BeginSearchLibrary(c.ID, me.ID); err != nil {
		t.Fatalf("beginning the search: %v", err)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("finding nothing: %v", err)
	}
	if n := searchEvents(g); n != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1: finding nothing is still a search", n)
	}
	shuffled := false
	for _, ev := range g.Events {
		if ev.Kind == EventSearchLibrary && ev.Label == SearchShuffledLabel {
			shuffled = true
		}
	}
	if !shuffled {
		t.Error("the search did not shuffle")
	}
}

// A mandatory search over no match looks, to every other seat, like one
// over a match: it asks, and its answer is "find nothing".
func TestZeroMatchMandatorySearchStillAsks(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Island")
	spec := optionalSwampSearch(me, nil)
	spec.Optional = false
	runSearch(t, g, spec)
	c := openSearch(g, me.ID)
	if c == nil || len(c.SearchCards) != 0 {
		t.Fatalf("a mandatory search with no match should ask with no candidates: %+v", c)
	}
	if c.SearchIsOptional() {
		t.Error("a mandatory search reports itself optional")
	}
	if err := g.DeclineSearchLibrary(c.ID, me.ID); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("declining a mandatory search = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("finding nothing: %v", err)
	}
	if n := searchEvents(g); n != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1", n)
	}
}
