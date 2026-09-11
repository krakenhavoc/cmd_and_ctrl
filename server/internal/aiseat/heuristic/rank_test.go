package heuristic_test

import (
	"context"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// rank_test.go covers the entry point Layer C needs: the prices
// Decide computes and throws away, so the layer above can see how
// close a call was.

func rankSeats() []protocol.PlayerView {
	return []protocol.PlayerView{
		newSeat(0, withHand(
			land("00000000-0000-0000-0000-0000000000a1", 0),
			creature("00000000-0000-0000-0000-0000000000a2", 0, "Bear", 2, 2),
		)),
		newSeat(1),
	}
}

// Rank prices every move and puts the best first. The land drop is
// the heuristic's highest-valued ordinary play, so it has to lead.
func TestRankPricesEveryMoveBestFirst(t *testing.T) {
	p := heuristic.New()
	v := newView(rankSeats(), withTurn(3, 0, "precombat_main"))
	in := input(0, v,
		passMove(0),
		castMove(t, 0, "00000000-0000-0000-0000-0000000000a2", "Cast Bear"),
		landMove(t, 0, "00000000-0000-0000-0000-0000000000a1"),
	)

	got := p.Rank(context.Background(), in)
	if len(got) != len(in.Moves) {
		t.Fatalf("ranked %d of %d moves", len(got), len(in.Moves))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Value < got[i].Value {
			t.Fatalf("ranking is not descending: %+v", got)
		}
	}
	if got[0].Index != 2 {
		t.Errorf("best move is index %d (%q); the land drop at 2 should lead",
			got[0].Index, in.Moves[got[0].Index].Label)
	}
	// Every index has to be a real move, and each exactly once.
	seen := map[int]bool{}
	for _, c := range got {
		if c.Index < 0 || c.Index >= len(in.Moves) {
			t.Fatalf("ranked an index that is not a move: %d", c.Index)
		}
		if seen[c.Index] {
			t.Fatalf("index %d ranked twice", c.Index)
		}
		seen[c.Index] = true
	}
	// Passing is the zero of the scale; the whole close-call trigger
	// is a comparison against it.
	for _, c := range got {
		if in.Moves[c.Index].Kind == "pass" && c.Value != 0 {
			t.Errorf("pass priced at %v, want 0", c.Value)
		}
	}
}

// Rank agrees with Decide about which move is best, or the escalation
// trigger is reading a different scorer than the one that plays.
func TestRankTopCandidateIsWhatDecideTakes(t *testing.T) {
	p := heuristic.New()
	v := newView(rankSeats(), withTurn(3, 0, "precombat_main"))
	in := input(0, v,
		passMove(0),
		castMove(t, 0, "00000000-0000-0000-0000-0000000000a2", "Cast Bear"),
		landMove(t, 0, "00000000-0000-0000-0000-0000000000a1"),
	)
	ranked := p.Rank(context.Background(), in)
	d, err := p.Decide(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if len(ranked) == 0 {
		t.Fatal("no ranking")
	}
	if d.Index != ranked[0].Index {
		t.Errorf("Decide took %d (%q) but Rank's best is %d (%q)",
			d.Index, in.Moves[d.Index].Label, ranked[0].Index, in.Moves[ranked[0].Index].Label)
	}
}

// The windows the scalar scorer does not answer must return nil
// rather than a fiction. A caller comparing the top two of a fiction
// escalates on noise, which costs money for nothing.
func TestRankDeclinesWindowsItDoesNotPrice(t *testing.T) {
	p := heuristic.New()
	atk := "00000000-0000-0000-0000-0000000000b1"
	blk := "00000000-0000-0000-0000-0000000000b2"
	v := newView(rankSeats(),
		withBattlefield(
			creature(atk, 0, "Bear", 2, 2),
			creature(blk, 1, "Ogre", 4, 4),
		),
		withTurn(3, 0, "declare_attackers"))

	if got := p.Rank(context.Background(), input(0, v, passMove(0), attackMove(t, 0, atk, 1))); got != nil {
		t.Errorf("an attack window was ranked: %+v", got)
	}
	if got := p.Rank(context.Background(), input(1, v, blockMove(t, 1, blk, atk))); got != nil {
		t.Errorf("a block window was ranked: %+v", got)
	}
	if got := p.Rank(context.Background(), input(0, newView(rankSeats()))); got != nil {
		t.Errorf("an empty window was ranked: %+v", got)
	}
}

// A choice window IS priced — those are the windows where the scorer
// picks the best of several answers, and a close call there is
// exactly what the model should be asked about.
func TestRankPricesChoiceWindows(t *testing.T) {
	p := heuristic.New()
	v := newView(rankSeats(), withTurn(3, 0, "precombat_main"))
	in := input(0, v,
		choiceMove(t, 0, "c1", "Discard Mountain", map[string]any{"card_ids": []string{"00000000-0000-0000-0000-0000000000a1"}}),
		choiceMove(t, 0, "c1", "Discard Bear", map[string]any{"card_ids": []string{"00000000-0000-0000-0000-0000000000a2"}}),
	)
	got := p.Rank(context.Background(), in)
	if len(got) != 2 {
		t.Fatalf("ranked %d of 2 choice answers", len(got))
	}
}

// Rank must respect ctx the same way decideGeneral does — it is
// called on the runner's clock.
func TestRankStopsOnACancelledContext(t *testing.T) {
	p := heuristic.New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	v := newView(rankSeats(), withTurn(3, 0, "precombat_main"))
	in := input(0, v, passMove(0), landMove(t, 0, "00000000-0000-0000-0000-0000000000a1"))
	if got := p.Rank(ctx, in); len(got) > len(in.Moves) {
		t.Errorf("ranked %d moves from a cancelled context", len(got))
	}
}
