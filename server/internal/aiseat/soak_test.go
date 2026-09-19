package aiseat_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soakPolicy builds the seat's policy for the soak run, per
// AISEAT_SOAK_POLICY:
//
//	random     (default) four random bots — the engine fuzzer
//	heuristic  four heuristic bots — S31 sub-PR 6's regression run
//	mixed      seats 0 and 2 heuristic, 1 and 3 random
//
// The mixed table is the interesting one for bug-hunting: a heuristic
// bot builds an orderly board and a random bot does something
// deranged to it, which reaches states neither policy gets to alone.
func soakPolicy(seed uint64, seat int) aiseat.Policy {
	switch os.Getenv("AISEAT_SOAK_POLICY") {
	case "heuristic":
		return heuristic.New()
	case "mixed":
		if seat%2 == 0 {
			return heuristic.New()
		}
	}
	return aiseat.NewRandomPolicy(rand.NewPCG(seed, uint64(seat)))
}

// TestRandomBotSoak is the opt-in engine fuzzer: N four-bot games,
// each to a winner or a turn budget, failing on the first stall.
// Skipped unless AISEAT_SOAK_GAMES is set. Its first 60 games found
// four engine bugs (turn rotation into eliminated seats, per-turn
// caches surviving the cleanup wrap, the damage-assignment
// prefix-lethal check on zero entries, and prompts owed by eliminated
// players); run it after any change to turn structure, combat, or
// pending choices.
//
//	AISEAT_SOAK_GAMES=200 go test ./internal/aiseat/ -run TestRandomBotSoak -v
//	AISEAT_SOAK_GAMES=100 AISEAT_SOAK_POLICY=heuristic go test ./internal/aiseat/ -run TestRandomBotSoak -v
//
// The name is historical: AISEAT_SOAK_POLICY chooses which policy
// fills the seats, and the heuristic and mixed tables run through the
// same harness rather than a parallel one.
func TestRandomBotSoak(t *testing.T) {
	n, _ := strconv.Atoi(os.Getenv("AISEAT_SOAK_GAMES"))
	if n <= 0 {
		t.Skip("set AISEAT_SOAK_GAMES=<n> to run the bot soak")
	}
	base := uint64(time.Now().UnixNano())
	if s, err := strconv.ParseUint(os.Getenv("AISEAT_SOAK_SEED"), 10, 64); err == nil {
		base = s
	}
	policy := os.Getenv("AISEAT_SOAK_POLICY")
	if policy == "" {
		policy = "random"
	}
	// #685: the per-game budgets were literals tuned on an idle
	// machine, which is the shape that took the nightly red in #600.
	// They are the same two knobs every other bot-table test reads, so
	// the nightly raises them once for the whole job. The defaults are
	// unchanged, so a local `AISEAT_SOAK_GAMES=200 go test` behaves
	// exactly as it did.
	stall := envDuration("AISEAT_STALL", 3*time.Second)
	wall := envDuration("AISEAT_WALLCLOCK", 60*time.Second)
	// One line naming everything a failure needs to be reproduced,
	// printed before the first game so it survives in the artifact of
	// a green run too.
	t.Logf("soak: %d %s games from seed %d (stall %s, wall %s); reproduce one with AISEAT_SOAK_GAMES=1 AISEAT_SOAK_SEED=<seed> AISEAT_SOAK_POLICY=%s",
		n, policy, base, stall, wall, policy)
	for i := 0; i < n; i++ {
		seed := base + uint64(i)
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			room := newRoom(t, 4, seed)
			g := room.Game
			ctx, cancel := context.WithTimeout(context.Background(), wall)
			defer cancel()
			var runners []*aiseat.Runner
			for i, p := range g.Seats {
				runners = append(runners, aiseat.Start(ctx, room, p.ID, soakPolicy(seed, i), aiseat.Config{}, nil, testLogger()))
			}
			lastSeq, lastMove := room.Seq(), time.Now()
			for {
				snap := g.Snapshot()
				if snap.State != game.StateActive || snap.Turn.Number > 80 {
					break
				}
				if seq := room.Seq(); seq != lastSeq {
					lastSeq, lastMove = seq, time.Now()
				} else if time.Since(lastMove) > stall {
					t.Fatalf("STALL (no seq movement in %s; reproduce with AISEAT_SOAK_SEED=%d AISEAT_SOAK_GAMES=1) turn %d step %s prio=%d pending=%d\n%s",
						stall, seed, snap.Turn.Number, snap.Turn.Step, snap.Turn.PriorityHolder, len(g.PendingChoices), describeSeats(g))
				}
				if ctx.Err() != nil {
					t.Fatalf("wall clock exhausted at turn %d (seed %d)", snap.Turn.Number, seed)
				}
				time.Sleep(5 * time.Millisecond)
			}
			cancel()
			for _, r := range runners {
				<-r.Done()
				for _, rej := range r.Stats().Rejections {
					if !isStepRace(rej) {
						t.Errorf("seed %d: enumerator offered a move the engine refused: %s %q: %v", seed, rej.Type, rej.Label, rej.Err)
					}
				}
			}
		})
	}
}
