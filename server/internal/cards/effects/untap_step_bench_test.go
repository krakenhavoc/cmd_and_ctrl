package effects

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_step_bench_test.go prices the change #74 made: the untap step
// went from silent to emitting one EventUntapCard per permanent it
// untaps, and every emit runs the S19 trigger harvester over the
// whole battlefield.
//
// That is the cost worth measuring, because it is the one that scales
// with the wrong thing. A four-player board can carry sixty
// permanents and untap twenty of them in a step, and the harvester's
// per-event walk is linear in board size — so the step's work is
// O(untapped x battlefield) where before it was O(battlefield).
//
// Two things keep the constant small, and both are visible in the
// numbers below:
//
//   - Only permanents that were actually TAPPED emit. CR 701.26b is a
//     change of state, so a board sitting upright costs nothing new
//     at all. The "nothing tapped" case is the old cost.
//   - The harvester's inner loop is a map lookup per battlefield card
//     and exits before allocating for any card with no declared
//     triggers, which on a real board is nearly all of them.
//
// Run: go test ./internal/cards/effects/ -run XXX -bench UntapStep -benchmem
//
// The catalog is live here — this package's init registers every
// card — so these are the real hook, the real registry and the real
// harvester, not a stand-in.

// benchBoard seats four players and parks `per` vanilla permanents on
// the battlefield for each, plus whatever `extra` names for seat 0.
// Returns the game and seat 0's permanents.
func benchBoard(b *testing.B, per int, extra ...[3]string) (*game.Game, uuid.UUID, []uuid.UUID) {
	b.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 40)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			b.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		b.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			b.Fatalf("KeepHand: %v", err)
		}
	}
	var mine []uuid.UUID
	for si, p := range g.Seats {
		for i := 0; i < per; i++ {
			id := uuid.New()
			g.Battlefield.PushTop(game.Card{
				InstanceID: id, Name: "Bench Rock", TypeLine: "Artifact",
				Owner: p.ID, Controller: p.ID,
			})
			if si == 0 {
				mine = append(mine, id)
			}
		}
	}
	for _, e := range extra {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: e[0], OracleID: e[1], TypeLine: e[2],
			Owner: g.Seats[0].ID, Controller: g.Seats[0].ID,
		})
	}
	return g, g.Seats[0].ID, mine
}

// benchUntap runs one untap of `mine` per iteration, re-tapping under
// a stopped timer. UntapAll is the public mutation over the same
// primitive and the same emits the untap step uses, so the per-event
// cost it measures is the step's.
func benchUntap(b *testing.B, g *game.Game, seat uuid.UUID, mine []uuid.UUID, tap bool) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if tap {
			for _, id := range mine {
				if err := g.TapCard(id, true); err != nil {
					b.Fatalf("TapCard: %v", err)
				}
			}
		}
		g.PendingTriggers = nil
		b.StartTimer()
		if err := g.UntapAll(seat); err != nil {
			b.Fatalf("UntapAll: %v", err)
		}
	}
}

// The baseline: a fifteen-permanent-per-seat board (sixty cards) with
// nothing tapped. No permanent becomes untapped, so no event is
// emitted and the step costs exactly what it cost before #74.
func BenchmarkUntapStepNothingTapped(b *testing.B) {
	g, seat, mine := benchBoard(b, 15)
	benchUntap(b, g, seat, mine, false)
}

// The same board with all fifteen of the active seat's permanents
// tapped: fifteen emits, each running the harvester over all sixty
// battlefield cards. This is the honest worst ordinary case.
func BenchmarkUntapStepFifteenTapped(b *testing.B) {
	g, seat, mine := benchBoard(b, 15)
	benchUntap(b, g, seat, mine, true)
}

// A deliberately outsized board — thirty per seat, a hundred and
// twenty cards, all thirty of the active seat's tapped — to show the
// quadratic term is still small in absolute terms at a size real
// games do not reach.
func BenchmarkUntapStepLargeBoardThirtyTapped(b *testing.B) {
	g, seat, mine := benchBoard(b, 30)
	benchUntap(b, g, seat, mine, true)
}

// The same fifteen-tapped board with a Mesmeric Orb on it: every
// untap now also matches a declared trigger, runs its Build and
// queues a StackItem. The delta against BenchmarkUntapStepFifteenTapped
// is what the card itself costs, as opposed to what observability
// costs.
func BenchmarkUntapStepFifteenTappedWithMesmericOrb(b *testing.B) {
	g, seat, mine := benchBoard(b, 15, [3]string{"Mesmeric Orb", mesmericOrbOracle, "Artifact"})
	benchUntap(b, g, seat, mine, true)
}
