package effects

// bench_test.go — the per-event cost of the card hooks on an
// 80-permanent board, committed from the September 2026 review (#561)
// so a change to the harvester, the layer engine or the replacement
// gather has a number a reviewer can ask for. Not a CI gate.
//
//	go test -run '^$' -bench . -benchmem ./internal/cards/effects/

import (
	"math/rand/v2"
	"testing"
	"unsafe"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// benchGame mirrors newCatalogGame in cards_test.go without the *testing.T.
func benchGame(b *testing.B) *game.Game {
	b.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			b.Fatal(err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		b.Fatal(err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			b.Fatal(err)
		}
	}
	return g
}

var benchCards = []struct{ name, oracle, tl string }{
	{"Soul Warden", "f3fad295-1af2-4ecc-8546-b121ad6be27b", "Creature — Human Cleric"},
	{"Impact Tremors", "9242cd3e-1a71-4700-8182-9c1005616033", "Enchantment"},
	{"Serra Angel", "4b7ac066-e5c7-43e6-9e7e-2739b24a905d", "Creature — Angel"},
	{"Glorious Anthem", "e3886fe8-9b76-4613-8891-4ec74657c087", "Enchantment"},
	{"Doubling Season", "01546b7d-a233-4176-8843-d732074dc5b6", "Enchantment"},
	{"Sol Ring", "6ad8011d-3471-4369-9d68-b264cc027487", "Artifact"},
	{"Llanowar Elves", "68954295-54e3-4303-a6bc-fc4547a4e3a3", "Creature — Elf Druid"},
	{"Grizzly Bears", "", "Creature — Bear"},
}

func seedBoard(g *game.Game, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		d := benchCards[i%len(benchCards)]
		owner := g.Seats[i%4].ID
		c := game.NewCard(d.name, owner)
		c.OracleID = d.oracle
		c.TypeLine = d.tl
		c.Controller = owner
		ids = append(ids, pushBattlefieldCardWithTimestamp(g, c))
	}
	return ids
}

func BenchmarkSizes(b *testing.B) {
	b.Logf("sizeof Spec=%d Card=%d Event=%d registry=%d", unsafe.Sizeof(Spec{}), unsafe.Sizeof(game.Card{}), unsafe.Sizeof(game.Event{}), len(registry))
}

func BenchmarkHarvestTapEvent_80perms(b *testing.B) {
	g := benchGame(b)
	ids := seedBoard(g, 80)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventTapCard, CardID: ids[i%80]})
		})
	}
}

func BenchmarkHarvestETBEvent_80perms(b *testing.B) {
	g := benchGame(b)
	ids := seedBoard(g, 80)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventETB, CardID: ids[i%80], Actor: g.Seats[i%4].ID})
			g.PendingTriggers = g.PendingTriggers[:0]
		})
	}
}

func BenchmarkLayerRecompute_80perms(b *testing.B) {
	g := benchGame(b)
	seedBoard(g, 80)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.WithWriteLock(func() {
			g.BumpLayerVersionForTest()
			g.RecomputeLayersIfStaleLocked()
		})
	}
}

func BenchmarkCounterWithReplacementGather_80perms(b *testing.B) {
	g := benchGame(b)
	ids := seedBoard(g, 80)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.WithWriteLock(func() {
			_ = g.AddCounterForEffect(ids[2], "+1/+1", 1)
		})
	}
}

func BenchmarkTallyScan_3000events(b *testing.B) {
	g := benchGame(b)
	p := g.Seats[0].ID
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: p})
		for i := 0; i < 3000; i++ {
			g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: p, Amount: 1})
		}
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = b15LifeGainedThisTurn(g, p)
	}
}
