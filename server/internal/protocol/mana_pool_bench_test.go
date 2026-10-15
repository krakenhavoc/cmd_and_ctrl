package protocol

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_pool_bench_test.go is #3041's measurement: what one action costs
// a table whose seat has a big floating pool, the shape a Marwyn at a
// million power makes in one tap. Each action takes an undo entry
// (Game.Clone), may write a restore point (CaptureSnapshot + JSON) and
// sends every viewer a frame (ViewOfGameFor + JSON); the legal-move
// lists ask the pool whether each card is payable (CanPayFor).

// bigPoolTable is landTable with `n` {G} from one Marwyn floating in
// the first seat's pool, added one mana at a time as the engine adds
// them.
func bigPoolTable(tb testing.TB, n int) (*game.Game, *game.Player) {
	tb.Helper()
	g := landTable(tb, 4, 0)
	p := g.Seats[0]
	marwyn := uuid.New()
	for range n {
		p.ManaPool.AddMana(game.ManaToken{Color: "G", Source: marwyn, SourceKinds: game.ManaSourceCreature})
	}
	return g, p
}

var bigPoolSizes = []int{1 << 16, 1 << 20}

func BenchmarkBigPoolClone(b *testing.B) {
	for _, n := range bigPoolSizes {
		b.Run(fmt.Sprintf("mana=%d", n), func(b *testing.B) {
			g, _ := bigPoolTable(b, n)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_ = g.Clone()
			}
		})
	}
}

func BenchmarkBigPoolRestorePoint(b *testing.B) {
	for _, n := range bigPoolSizes {
		b.Run(fmt.Sprintf("mana=%d", n), func(b *testing.B) {
			g, _ := bigPoolTable(b, n)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if _, err := json.Marshal(g.CaptureSnapshot()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkBigPoolView(b *testing.B) {
	for _, n := range bigPoolSizes {
		b.Run(fmt.Sprintf("mana=%d", n), func(b *testing.B) {
			g, p := bigPoolTable(b, n)
			viewer := p.ID.String()
			_ = ViewOfGameFor(g, viewer)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if _, err := json.Marshal(ViewOfGameFor(g, viewer)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkBigPoolCanPay(b *testing.B) {
	cost, err := game.ParseCost("{X}{G}{G}")
	if err != nil {
		b.Fatal(err)
	}
	for _, n := range bigPoolSizes {
		b.Run(fmt.Sprintf("mana=%d", n), func(b *testing.B) {
			_, p := bigPoolTable(b, n)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if !p.ManaPool.CanPayFor(cost, n-2, game.ManaSpendContext{}) {
					b.Fatal("a pool of exactly X+2 {G} could not pay {X}{G}{G}")
				}
			}
		})
	}
}
