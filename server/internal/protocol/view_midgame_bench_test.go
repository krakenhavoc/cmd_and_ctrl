package protocol

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// midgameTable is the board #3823 measures the per-commit view on, the
// protocol-side twin of game.BenchmarkCloneMidgame's (ADR 0194): four
// seats with 100-card decks, about 25 permanents a seat (lands with a
// mana ability, creatures, a third of them with counters), 15 cards in
// each graveyard, the opening hand, and a few thousand events in the
// log. The cards are moved between zones directly, which is fine for a
// benchmark of the projection and nothing else.
func midgameTable(tb testing.TB) *game.Game {
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
	if err := g.Start(rand.New(rand.NewPCG(3823, 1))); err != nil {
		tb.Fatalf("Start: %v", err)
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
	})
	growLongGame(tb, g, 5000)
	return g
}

// BenchmarkViewOfGameMidgame is the view ws.Room.captureLocked builds on
// every commit (#3823), on a realistic four-seat board.
func BenchmarkViewOfGameMidgame(b *testing.B) {
	g := midgameTable(b)
	_ = ViewOfGame(g)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = ViewOfGame(g)
	}
}

// BenchmarkCommitMidgame splits the rest of a production commit's
// per-frame work (#3823) on the same board, each part on its own, so
// the parts can be weighed against the view: the crash-recovery dump
// and replay line (the unfiltered view marshalled once), the restore
// point (the domain snapshot marshalled; the atomic file write is left
// out), and the broadcast (one filter and one marshal per seat).
func BenchmarkCommitMidgame(b *testing.B) {
	g := midgameTable(b)
	v := ViewOfGame(g)
	b.Run("dump_marshal", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			if _, err := json.Marshal(SnapshotPayload{Seq: 1, Game: v}); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("restore_point_marshal", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			if _, err := json.Marshal(g.CaptureSnapshot()); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("broadcast_four_seats", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			for _, p := range g.Seats {
				if _, err := json.Marshal(SnapshotPayload{Seq: 1, Game: FilterViewFor(v, p.ID.String())}); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}
