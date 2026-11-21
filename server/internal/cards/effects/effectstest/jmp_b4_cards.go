package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared --------------------------------------------------------------

// jmpB4CastCreature puts a catalog creature card in p's hand and casts it
// with the given params. The caller is at a point where p may cast it.
func JmpB4CastCreature(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// jmpB4Live is the permanent with its layers applied.
func JmpB4Live(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := FindBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c
}

func JmpB4PlusOneCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return JmpB4Live(t, g, id).Counters[game.CounterPlusOne]
}

// jmpB4Pick answers the latest pick_target prompt of chooser with the cards
// given (none is a legal answer to an "up to" clause).
func JmpB4Pick(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) error {
	t.Helper()
	pick := LatestPickTarget(g, chooser)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	var refs []game.TargetRef
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return g.ResolvePickTargets(pick.ID, chooser, refs)
}
