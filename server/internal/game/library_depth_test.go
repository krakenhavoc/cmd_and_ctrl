package game

import (
	"testing"

	"github.com/google/uuid"
)

// library_depth_test.go — "into its owner's library third from the
// top" (Teferi, Hero of Dominaria's −3), which is the first printed
// library POSITION in the engine that is neither the top nor the
// bottom.

func TestInsertFromTopPlacesTheCardAtTheNamedDepth(t *testing.T) {
	names := func(z *Zone) []string {
		out := make([]string, 0, len(z.Cards))
		// Top-first, which is how a player reads a library.
		for i := len(z.Cards) - 1; i >= 0; i-- {
			out = append(out, z.Cards[i].Name)
		}
		return out
	}
	fresh := func() *Zone {
		z := newZone(ZoneLibrary, uuid.New())
		// Pushed bottom-up, so "D" ends on top.
		for _, n := range []string{"A", "B", "C", "D"} {
			z.PushTop(Card{InstanceID: uuid.New(), Name: n})
		}
		return z
	}

	for _, tc := range []struct {
		depth int
		want  []string
	}{
		{1, []string{"X", "D", "C", "B", "A"}},
		{2, []string{"D", "X", "C", "B", "A"}},
		{3, []string{"D", "C", "X", "B", "A"}},
		{5, []string{"D", "C", "B", "A", "X"}},
		// Deeper than the library is tall: as close to the printed
		// position as the library can get, which is the bottom.
		{9, []string{"D", "C", "B", "A", "X"}},
		// A caller that has not decided gets the top.
		{0, []string{"X", "D", "C", "B", "A"}},
	} {
		z := fresh()
		z.InsertFromTop(Card{InstanceID: uuid.New(), Name: "X"}, tc.depth)
		got := names(z)
		if len(got) != len(tc.want) {
			t.Fatalf("depth %d: %v, want %v", tc.depth, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("depth %d: %v, want %v", tc.depth, got, tc.want)
			}
		}
	}
}

func TestInsertFromTopIntoAnEmptyZone(t *testing.T) {
	z := newZone(ZoneLibrary, uuid.New())
	id := uuid.New()
	z.InsertFromTop(Card{InstanceID: id, Name: "Only"}, 3)
	if z.Size() != 1 {
		t.Fatalf("size = %d, want 1", z.Size())
	}
	top, err := z.Top()
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if top.InstanceID != id {
		t.Error("the only card in the library is not on top")
	}
}

// TestTuckAtDepthLandsThirdFromTheTop is the route, not the zone
// helper: the card leaves the battlefield through the shared exit and
// arrives at the depth the route asked for.
func TestTuckAtDepthLandsThirdFromTheTop(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	if owner.Library.Size() < 3 {
		t.Fatal("test needs a library of at least three cards")
	}
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Tucked",
		TypeLine:   "Creature — Test",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	before := owner.Library.Size()

	g.mu.Lock()
	err := g.TuckToLibraryAtDepthForEffect(id, 3)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("TuckToLibraryAtDepthForEffect: %v", err)
	}
	if owner.Library.Size() != before+1 {
		t.Fatalf("library = %d, want %d", owner.Library.Size(), before+1)
	}
	if g.Battlefield.Contains(id) {
		t.Fatal("the permanent is still on the battlefield")
	}
	top := owner.Library.Size() - 1
	if owner.Library.Cards[top-2].InstanceID != id {
		t.Error("the tucked permanent is not third from the top")
	}
}

// TestCommanderTuckedAtDepthDeclineKeepsTheDepth is the pair to
// TestCommanderTuckedToBottomDeclineKeepsTheBottom: the depth is the
// second piece of per-route bookkeeping that has to survive the CR
// 903.9 pause, and it survives for the same reason — it rides the
// route rather than being applied by the caller.
func TestCommanderTuckedAtDepthDeclineKeepsTheDepth(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	if owner.Library.Size() < 3 {
		t.Fatal("test needs a library of at least three cards")
	}
	cmdID := seatCommander(t, g.Battlefield, owner)

	g.mu.Lock()
	err := g.TuckToLibraryAtDepthForEffect(cmdID, 3)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("TuckToLibraryAtDepthForEffect: %v", err)
	}
	if owner.Library.Contains(cmdID) {
		t.Fatal("the commander hit the library before the prompt was answered")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Library, owner.Command, g.Battlefield)
	top := owner.Library.Size() - 1
	if owner.Library.Cards[top-2].InstanceID != cmdID {
		t.Error("the depth did not survive the CR 903.9 pause")
	}
}

// TestCommanderTuckedAtDepthAcceptGoesToTheCommandZone — the other
// answer, and the one that matters at a real table: a tucked
// commander is not lost in a library unless its owner chooses that.
func TestCommanderTuckedAtDepthAcceptGoesToTheCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)

	g.mu.Lock()
	err := g.TuckToLibraryAtDepthForEffect(cmdID, 3)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("TuckToLibraryAtDepthForEffect: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Library, g.Battlefield)
}
