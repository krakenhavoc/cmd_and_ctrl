package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// search_fail_to_find_test.go — #3053. The enumerator is the closed move
// list a bot picks from, so "find nothing" has to be on it exactly when
// the engine would accept it (CR 701.23b), and absent exactly when it
// would not (CR 701.23d). Both directions are checked against the
// dispatcher with dispatchAll, the package's one promise.

func stockedLibrary(t *testing.T, p *game.Player, names ...string) []uuid.UUID {
	t.Helper()
	p.Library.Cards = nil
	ids := make([]uuid.UUID, 0, len(names))
	for _, n := range names {
		c := basic(n, n)
		c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
		p.Library.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func openSearchPrompt(t *testing.T, g *game.Game, spec game.SearchLibrarySpec) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.SearchLibraryThenForEffect(spec); err != nil {
			t.Fatalf("SearchLibraryThenForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceSearchLibrary {
		t.Fatalf("expected one search prompt, got %+v", g.PendingChoices)
	}
}

func choiceMoves(moves []legal.Move) (failToFind []legal.Move, takes []legal.Move) {
	for _, m := range moves {
		if m.Kind != legal.KindChoice {
			continue
		}
		if strings.Contains(m.Label, "fail to find") {
			failToFind = append(failToFind, m)
		} else {
			takes = append(takes, m)
		}
	}
	return failToFind, takes
}

// One match for a stated quality: both answers are on offer, and the
// fail-to-find is the always-legal one.
func TestOneMatchSearchEnumeratesFailToFindAndTheTake(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	ids := stockedLibrary(t, active, "Island", "Forest", "Swamp")
	openSearchPrompt(t, g, game.SearchLibrarySpec{
		Player: active.ID, Pred: func(c game.Card) bool { return c.Name == "Forest" },
		Dest: game.ZoneHand, Limit: 1, Shuffle: true,
	})

	moves := legal.EnumerateFor(g, active.ID)
	fail, takes := choiceMoves(moves)
	if len(fail) != 1 {
		t.Fatalf("want exactly one fail-to-find move for a one-match search, got %v", labels(moves))
	}
	if !fail[0].AlwaysLegal {
		t.Error("fail to find is the choice's always-legal answer when the floor is zero")
	}
	if len(takes) != 1 || len(pickedCardIDs(t, takes[0])) != 1 || pickedCardIDs(t, takes[0])[0] != ids[1] {
		t.Fatalf("want exactly one take, of the Forest, got %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// "Up to two" against exactly two matches: none, each single, and the pair.
func TestUpToTwoSearchEnumeratesEveryCount(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	stockedLibrary(t, active, "Forest", "Forest", "Island")
	openSearchPrompt(t, g, game.SearchLibrarySpec{
		Player: active.ID, Pred: func(c game.Card) bool { return c.Name == "Forest" },
		Dest: game.ZoneHand, Limit: 2, Shuffle: true,
	})

	moves := legal.EnumerateFor(g, active.ID)
	fail, takes := choiceMoves(moves)
	sizes := map[int]int{}
	for _, m := range takes {
		sizes[len(pickedCardIDs(t, m))]++
	}
	// The two Forests are interchangeable, so taking one is one move
	// (interchangeableSearchPool), and taking both is the other.
	if len(fail) != 1 || sizes[1] != 1 || sizes[2] != 1 {
		t.Fatalf("want fail-to-find, one single and one pair; got %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// Demonic Tutor: no stated quality, so there is no fail-to-find move at
// all, and the dispatcher would refuse it.
func TestUnconditionalSearchDoesNotEnumerateFailToFind(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	stockedLibrary(t, active, "Island", "Forest", "Swamp")
	openSearchPrompt(t, g, game.SearchLibrarySpec{
		Player: active.ID, Dest: game.ZoneHand, Limit: 1, Shuffle: true,
	})

	moves := legal.EnumerateFor(g, active.ID)
	fail, takes := choiceMoves(moves)
	if len(fail) != 0 {
		t.Fatalf("an unconditional search must find: fail to find was offered: %v", labels(moves))
	}
	if len(takes) != 3 {
		t.Fatalf("want a take per card, got %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// "Three cards": every offered pick has three.
func TestThreeCardUnconditionalSearchOnlyOffersThrees(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	stockedLibrary(t, active, "A", "B", "C", "D")
	openSearchPrompt(t, g, game.SearchLibrarySpec{
		Player: active.ID, Dest: game.ZoneHand, Limit: 3, Shuffle: true,
	})

	moves := legal.EnumerateFor(g, active.ID)
	fail, takes := choiceMoves(moves)
	if len(fail) != 0 || len(takes) == 0 {
		t.Fatalf("want only takes, got %v", labels(moves))
	}
	for _, m := range takes {
		if n := len(pickedCardIDs(t, m)); n != 3 {
			t.Errorf("offered a %d-card pick for a three-card search: %q", n, m.Label)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

// "Up to three cards", no quality: fewer is allowed again.
func TestUpToUnconditionalSearchEnumeratesFailToFind(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	stockedLibrary(t, active, "A", "B", "C", "D")
	openSearchPrompt(t, g, game.SearchLibrarySpec{
		Player: active.ID, Dest: game.ZoneHand, Limit: 3, UpTo: true, Shuffle: true,
	})
	moves := legal.EnumerateFor(g, active.ID)
	fail, _ := choiceMoves(moves)
	if len(fail) != 1 {
		t.Fatalf("up to N may take none: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}
