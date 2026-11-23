package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_alt_cost_cards_test.go — ADR 0135 §2 (#2412): alternative costs
// that discard cards, from hand, against the real catalog. Snag, Abolish,
// Flameshot and Outbreak discard one card of a land type; Foil discards
// an Island card and another card under a set rule; The Infamous
// Cruelclaw grants "cast it by discarding a card".

const (
	SnagOracle       = "feb5e9a9-500a-4f20-881d-c97d052953ce"
	AbolishOracle    = "d6d7faa8-ecc9-408b-9c46-a0f62c74f567"
	FlameshotOracle  = "26640648-6400-4b80-a78b-f519ab9cd09e"
	OutbreakOracle   = "05079cee-b190-42d7-b59f-130cb545cab7"
	FoilOracle       = "7dde4eb6-e9d7-4259-abc2-3af738e0f00f"
	CruelclawOracle  = "e229d55c-9513-437d-bd67-015ac46aaa18"
	DiscardAltCostKy = "discard"
)

// discardTable is a main phase on seat 0's turn with an empty hand.
func DiscardTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me, g.Seats[1]
}

func HandCardOf(g *game.Game, p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true}})
	})
	return id
}
