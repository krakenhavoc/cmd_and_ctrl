package game

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
)

// layers_concurrency_test.go is the regression suite for the layer
// engine's lock contract.
//
// The recompute pass WRITES: `Card.effective` on every battlefield
// card, and — since the S24 layer-2 control change —
// `Card.Controller`, `Card.SummonedThisTurn`, `Card.AttackingTarget`
// and `Card.BlockingTarget` on any permanent whose controller just
// changed. Those writes have to be exclusive against every other
// reader of the same fields.
//
// Before the fix, `RecomputeLayersIfStaleLocked` claimed it was safe
// under the game's READ lock: version reads/writes are atomic and a
// dedicated mutex serialised recomputes against each other. That
// serialises WRITERS against WRITERS and says nothing about a reader
// holding only the read lock — and `sync.RWMutex` lets any number of
// those run at once. So a snapshot that recomputed raced every
// concurrent read-lock holder.
//
// The exposure is not hypothetical wiring. Two production callers
// take the read lock on the same *Game from different goroutines:
//
//   - `protocol.ViewOfGame` → `Game.ReadSnapshot` (the WS broadcast
//     goroutine), which is the caller that runs the recompute.
//   - `Game.AutoTapForCostExcluding`, reached from the lobby's
//     `/games/{id}/autotap` HTTP handler on a net/http goroutine; it
//     reads `c.Controller` off every battlefield card
//     (autotap.go: `if c.Controller != controller || c.Tapped`).
//     `Game.ControllerOfCard` (the actions package's per-card
//     authorization gate) reads the same field the same way.
//
// A torn `Card.Controller` is a 16-byte uuid.UUID read that can mix
// halves of two different players' IDs — a permanent silently
// attributed to the wrong seat, which is the "my opponent attacked me
// with my own creature" report nobody can reproduce.
//
// The tests below run those exact pairings under `go test -race`.
// They fail (as race reports, not assertion failures) on the old
// contract and pass on the new one, where the recompute runs under
// the WRITE lock.

// flippingControlStatic returns a layer-2 control-changer whose
// AppliesTo is gated on an atomic the test owns. Toggling the gate
// between recomputes makes the materialisation write
// `Card.Controller` on EVERY pass rather than only on the first,
// which turns a rare interleaving into one the race detector sees
// immediately.
func flippingControlStatic(victim uuid.UUID, on *atomic.Bool) StaticAbility {
	return StaticAbility{
		Layer: Layer2Control,
		AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
			return on.Load() && target.InstanceID == victim
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
			c.Controller = source.Controller
		},
	}
}

// raceFixture seeds a two-seat game with a permanent that seat 1
// owns and a "thief" seat 0 controls, wired to a layer-2 static that
// the returned atomic switches on and off.
func raceFixture(t *testing.T) (g *Game, victim uuid.UUID, on *atomic.Bool) {
	t.Helper()
	g = newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim = pushAttachTestCard(g, opp.ID, "Bear", "Creature — Bear")
	thief := pushAttachTestCard(g, me.ID, "Thief", "Enchantment")
	// A couple of extra permanents so the recompute has real work to
	// do and the write window is not a single instruction wide.
	for i := 0; i < 6; i++ {
		pushAttachTestCard(g, me.ID, "Forest", "Basic Land — Forest")
	}

	on = &atomic.Bool{}
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != "thief" {
			return nil
		}
		return []StaticAbility{flippingControlStatic(victim, on)}
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == thief {
				g.Battlefield.Cards[i].OracleID = "thief"
			}
		}
	})
	return g, victim, on
}

// TestRecomputeDoesNotRaceControllerOfCard pins the actions-package
// pairing: a snapshot that recomputes layers against the per-card
// authorization gate, both on their own goroutines, both previously
// holding nothing stronger than the read lock.
func TestRecomputeDoesNotRaceControllerOfCard(t *testing.T) {
	g, victim, on := raceFixture(t)

	const rounds = 300
	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Readers: g.ControllerOfCard takes the read lock and reads
	// Card.Controller straight off the battlefield.
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				g.ControllerOfCard(victim)
			}
		}()
	}

	// Recomputer: exactly what the WS broadcast goroutine does.
	for i := 0; i < rounds; i++ {
		on.Store(i%2 == 0)
		g.BumpLayerVersionForTest()
		g.ReadSnapshot(func() {})
	}
	close(stop)
	wg.Wait()
}

// TestRecomputeDoesNotRaceAutoTap pins the lobby pairing: the
// /autotap HTTP preview walks the battlefield under the read lock
// while the broadcast goroutine recomputes. AutoTapForCostExcluding
// reads Card.Controller and the effective characteristic, so it
// covers both the S24 control materialisation and the pre-existing
// `c.effective = &printed` write.
func TestRecomputeDoesNotRaceAutoTap(t *testing.T) {
	g, _, on := raceFixture(t)
	seat := g.Seats[0].ID
	cost, err := ParseCost("{1}{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}

	const rounds = 300
	var wg sync.WaitGroup
	stop := make(chan struct{})

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				g.AutoTapForCostExcluding(seat, cost, 0, nil)
			}
		}()
	}

	for i := 0; i < rounds; i++ {
		on.Store(i%2 == 0)
		g.BumpLayerVersionForTest()
		g.ReadSnapshot(func() {})
	}
	close(stop)
	wg.Wait()
}

// TestRecomputeDoesNotRaceEffectiveReads covers the OLDER half of
// the same exposure: `c.effective = &printed` runs for every
// battlefield card on every pass, so a snapshot that recomputes
// raced any concurrent snapshot reading an effective characteristic
// — power/toughness, types, granted keywords, all of it. That
// predates S24's control change; the control change only made the
// consequence silent instead of a nil-pointer crash. Both are fixed
// by the same move of the recompute under the write lock, which is
// why they are pinned together rather than split across PRs.
func TestRecomputeDoesNotRaceEffectiveReads(t *testing.T) {
	g, _, on := raceFixture(t)

	const rounds = 300
	var wg sync.WaitGroup
	stop := make(chan struct{})

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				g.ReadSnapshot(func() {
					for _, c := range g.Battlefield.Cards {
						_ = c.CurrentPower()
						_ = c.IsCreature()
					}
				})
			}
		}()
	}

	for i := 0; i < rounds; i++ {
		on.Store(i%2 == 0)
		g.BumpLayerVersionForTest()
		g.ReadSnapshot(func() {})
	}
	close(stop)
	wg.Wait()
}

// TestConcurrentSnapshotsAgree is the correctness half: whatever the
// locking, two goroutines both snapshotting must never observe a
// controller that is neither the baseline nor the thief. It also
// exercises two recomputers against each other, which the old
// recompute mutex did cover — the new contract must not lose that.
func TestConcurrentSnapshotsAgree(t *testing.T) {
	g, victim, on := raceFixture(t)
	on.Store(true)
	baseline, thiefSeat := g.Seats[1].ID, g.Seats[0].ID

	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				g.BumpLayerVersionForTest()
				var got uuid.UUID
				g.ReadSnapshot(func() {
					for _, c := range g.Battlefield.Cards {
						if c.InstanceID == victim {
							got = c.Controller
						}
					}
				})
				if got != baseline && got != thiefSeat {
					t.Errorf("controller %s is neither baseline %s nor thief %s",
						got, baseline, thiefSeat)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// BenchmarkReadSnapshotFresh measures the fast path — the one
// ViewOfGame hits on every broadcast where nothing invalidated the
// layer cache. It must stay a lock acquisition plus two atomic
// loads.
func BenchmarkReadSnapshotFresh(b *testing.B) {
	g := benchGame(b)
	g.ReadSnapshot(func() {})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.ReadSnapshot(func() {})
	}
}

// BenchmarkReadSnapshotStale measures the slow path: every iteration
// invalidates the cache so the snapshot pays for a full recompute.
// This is the number the lock change could regress.
func BenchmarkReadSnapshotStale(b *testing.B) {
	g := benchGame(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.BumpLayerVersionForTest()
		g.ReadSnapshot(func() {})
	}
}

// BenchmarkReadSnapshotStaleContended is the shape that matters for
// the fix: a stale snapshot while other goroutines hold the read
// lock. Under the old contract they overlapped (and raced); under the
// new one the recompute excludes them.
func BenchmarkReadSnapshotStaleContended(b *testing.B) {
	g := benchGame(b)
	// A real battlefield card, so the readers' hold time is a
	// battlefield scan and not a walk of four 99-card libraries
	// looking for an ID that is not there.
	card := g.Battlefield.Cards[0].InstanceID
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				g.ControllerOfCard(card)
			}
		}()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.BumpLayerVersionForTest()
		g.ReadSnapshot(func() {})
	}
	b.StopTimer()
	close(stop)
	wg.Wait()
}

// benchGame seeds a battlefield roughly the size of a real four-way
// Commander board mid-game.
func benchGame(b *testing.B) *Game {
	b.Helper()
	g := NewGame()
	for i := 0; i < 4; i++ {
		if _, err := g.AddPlayer("P", buildTestDeck("Commander")); err != nil {
			b.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(nil); err != nil {
		b.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		for i := 0; i < 10; i++ {
			g.Battlefield.PushTop(Card{
				InstanceID: uuid.New(),
				Name:       "Bear",
				TypeLine:   "Creature — Bear",
				Owner:      p.ID,
				Controller: p.ID,
			})
		}
	}
	return g
}
