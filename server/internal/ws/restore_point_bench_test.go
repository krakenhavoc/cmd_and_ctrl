package ws

import (
	"fmt"
	"math/rand/v2"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// midgameRestoreTable is the board #3823 measured the restore point on,
// the ws twin of protocol's midgameTable: four seats with 100-card
// decks, about 25 permanents a seat (a third with counters and extra
// keywords), 15 cards in each graveyard, the opening hand, and `events`
// events in the log, in the mix protocol's growLongGame emits (a step
// every 40, a reveal every 150). Cards are moved between zones
// directly, which is fine for a benchmark of the capture and the
// encoding and nothing else.
func midgameRestoreTable(tb testing.TB, events int) *game.Game {
	tb.Helper()
	g := game.NewGame()
	for i := range 4 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 99 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat %d", i+1), deck); err != nil {
			tb.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithSource(rand.NewPCG(3823, 1)); err != nil {
		tb.Fatalf("StartWithSource: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			tb.Fatalf("KeepHand: %v", err)
		}
	}
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			for i := 0; i < 25 && len(p.Library.Cards) > 0; i++ {
				c := p.Library.Cards[0]
				p.Library.Cards = p.Library.Cards[1:]
				switch {
				case i < 10:
					c.Name, c.TypeLine = "Mountain", "Basic Land — Mountain"
				default:
					c.Name, c.TypeLine, c.ManaCost = fmt.Sprintf("Beast %d", i), "Creature — Beast", "{2}{R}"
					c.Power, c.Toughness = 3, 3
					if i%3 == 0 {
						c.Counters = map[string]int{"+1/+1": i}
						c.Keywords = append(c.Keywords, "flying", "trample")
					}
				}
				g.Battlefield.Cards = append(g.Battlefield.Cards, c)
			}
			for i := 0; i < 15 && len(p.Library.Cards) > 0; i++ {
				p.Graveyard.Cards = append(p.Graveyard.Cards, p.Library.Cards[0])
				p.Library.Cards = p.Library.Cards[1:]
			}
		}
		for i := 0; i < events; i++ {
			active := g.Seats[(i/40)%len(g.Seats)]
			switch {
			case i%40 == 0:
				g.EmitEvent(game.Event{
					Kind: game.EventStepBegan, Actor: active.ID, Amount: i/40 + 1,
					Label: string(game.StepPrecombatMain),
				})
			case i%150 == 0:
				g.RevealTopOfLibraryForEffect(active.ID, uuid.Nil, 3, "bench")
			default:
				g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: active.ID})
			}
		}
	})
	return g
}

// emitOne is a commit's worth of new history: one more event.
func emitOne(g *game.Game) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: g.Seats[0].ID})
	})
}

// BenchmarkRestorePointMidgame is #3823's restore point on the midgame
// board with 5,000 events.
//
//   - write: one restore point written (capture, encode, atomic file
//     write) after each one-event commit. Before the encoded-event
//     cache this re-encoded the whole log every time.
//   - commit: a whole room commit (view, crash dump, replay line,
//     restore point) of a one-event action, writing the restore point
//     on every commit.
//   - commit_rate_limited: the same at the server's default interval,
//     so a commit pays the capture and the write is due later.
func BenchmarkRestorePointMidgame(b *testing.B) {
	b.Run("write", func(b *testing.B) {
		g := midgameRestoreTable(b, 5000)
		r := NewRoom(g, nil, b.TempDir())
		b.ReportAllocs()
		b.ResetTimer()
		for i := range b.N {
			emitOne(g)
			r.mu.Lock()
			wrote, err := r.writeRestorePointLocked(uint64(i + 1))
			r.mu.Unlock()
			if err != nil || !wrote {
				b.Fatalf("write = %v, %v; the bench board must be a restore point", wrote, err)
			}
		}
	})
	for _, tc := range []struct {
		name  string
		every time.Duration
	}{{"commit", 0}, {"commit_rate_limited", DefaultRestorePointInterval}} {
		b.Run(tc.name, func(b *testing.B) {
			g := midgameRestoreTable(b, 5000)
			r := NewRoom(g, nil, b.TempDir())
			r.SetRestorePointInterval(tc.every)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if _, _, err := r.Apply(g.Seats[0].ID, func() error { emitOne(g); return nil }); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			r.FlushRestorePoint()
		})
	}
}

// TestEncodedEventLogMemory states what the encoded-event cache costs a
// table (#3823, ADR 0041's 2026-10-15 amendment): the bytes it holds
// per event and in all, on the midgame board at 5,000 events, beside
// the restore file it serves. Run with -v to read the numbers.
func TestEncodedEventLogMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a 5,000-event game")
	}
	g := midgameRestoreTable(t, 5000)
	dir := t.TempDir()
	r := NewRoom(g, nil, dir)
	r.mu.Lock()
	wrote, err := r.writeRestorePointLocked(1)
	r.mu.Unlock()
	if err != nil || !wrote {
		t.Fatalf("write = %v, %v", wrote, err)
	}
	st, err := os.Stat(restorePointPath(dir, g.ID))
	if err != nil {
		t.Fatal(err)
	}
	n, held := r.eventLog.Len(), r.eventLog.Bytes()
	t.Logf("events=%d cache=%d bytes (%.0f per event) restore_file=%d bytes", n, held, float64(held)/float64(n), st.Size())
	if n != len(g.Events) {
		t.Errorf("cache holds %d events, the log %d", n, len(g.Events))
	}
	// The cache is the log's encoding plus an offset a event, with at
	// most append's doubling on top: never more than three times the
	// file.
	if held > 3*int(st.Size()) {
		t.Errorf("cache holds %d bytes for a %d-byte restore file", held, st.Size())
	}
}
