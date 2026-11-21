package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2166: mana you don't lose as steps and phases end. The statics
// (Upwelling, Leyline Tyrant, Kruphix) are derived from the battlefield;
// the mark (Karn, Savage Ventmaw) and the granted statement (The Last
// Agni Kai) are per-mana and per-player state.

const (
	MkUpwellingOracle = "420ff0f4-6056-4b40-9a28-fd4c23d6b81c"
	MkTyrantOracle    = "f92aaa00-6ece-4033-b3df-b2fc2c4718d9"
	MkKruphixOracle   = "b2cefcd6-4b81-479c-86ff-1695b836972c"
	MkVentmawOracle   = "f743d85f-754e-4e70-9907-af54acb9e8d4"
	MkKarnOracle      = "bc8802db-8c94-4eb4-817f-5f4f5408b7a4"
	MkAgniKaiOracle   = "fd1998c0-c18e-4310-82de-b6320be5fe10"
)

func MkAdvanceTo(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if g.Turn.Step == step {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached %s", step)
}

func MkAdvance(t *testing.T, g *game.Game) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
}

func MkAdd(t *testing.T, g *game.Game, p *game.Player, produced string, opts game.AddManaOptions) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaWithOptionsForEffect(p.ID, uuid.Nil, produced, opts); err != nil {
			t.Fatalf("AddMana: %v", err)
		}
	})
}

func MkColors(p *game.Player) map[string]int {
	out := map[string]int{}
	for _, tok := range p.ManaPool.Units() {
		out[tok.Color]++
	}
	return out
}

func MkWantPool(t *testing.T, p *game.Player, want map[string]int) {
	t.Helper()
	got := MkColors(p)
	if len(got) != len(want) {
		t.Fatalf("pool = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("pool = %v, want %v", got, want)
		}
	}
}
