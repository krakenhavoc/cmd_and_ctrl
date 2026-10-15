package game

import (
	"testing"
	"unsafe"
)

// midgameBoard builds the shape #3136 measures Clone against: a
// four-seat Commander table some turns in, with most of each 99-card
// deck still in the library, a battlefield of about 25 permanents a
// seat (some with counters and keywords), a graveyard, a hand and a
// long event log. It moves cards between zones directly, which is
// fine for a benchmark of the copy and nothing else.
func midgameBoard(tb testing.TB) *Game {
	tb.Helper()
	g := newActiveGameWithSeats(tb, 4)
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			for i := 0; i < 25 && len(p.Library.Cards) > 0; i++ {
				c := p.Library.Cards[0]
				p.Library.Cards = p.Library.Cards[1:]
				if i%3 == 0 {
					c.Counters = map[string]int{"+1/+1": i}
					c.Keywords = append(c.Keywords, "flying", "trample")
				}
				g.Battlefield.Cards = append(g.Battlefield.Cards, c)
			}
			for i := 0; i < 15 && len(p.Library.Cards) > 0; i++ {
				p.Graveyard.Cards = append(p.Graveyard.Cards, p.Library.Cards[0])
				p.Library.Cards = p.Library.Cards[1:]
			}
		}
	})
	emitFillerEvents(g, 5_000)
	return g
}

// BenchmarkCloneMidgame is the per-action undo snapshot (ws.Room.apply)
// on a realistic board (ADR 0194). card-bytes is unsafe.Sizeof(Card{}):
// the bytes per Clone are about that times the cards on the table.
func BenchmarkCloneMidgame(b *testing.B) {
	g := midgameBoard(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Clone()
	}
	b.ReportMetric(float64(unsafe.Sizeof(Card{})), "card-bytes")
}
