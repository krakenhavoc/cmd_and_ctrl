package effectstest

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sideboard_cards_test.go — ADR 0191: the cards that bring a card in
// from outside the game, over a deck's sideboard.

const (
	ExtrapolateOracle = "1f47c869-f4a9-43b0-aec0-8da4e3a3fec5"
	BurningWishOracle = "13a884b5-8970-4d71-9082-a8b6c8db9a01"
	LivingWishOracle  = "b9d186f5-1c92-4539-a5b5-e7fbd4b726fe"
	MastermindsOracle = "6aec5b1f-cea6-4ec3-9759-e4544da8c52b"
)

func OutsideCard(name, typeLine string) game.Card {
	c := game.NewCard(name, uuid.Nil)
	c.TypeLine = typeLine
	c.IsSideboard = true
	return c
}

// newSideboardCatalogGame is newCatalogGame with `seats` seats, seat 0
// owning `side` outside the game.
func NewSideboardCatalogGame(t *testing.T, seats int, side ...game.Card) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < seats; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if i == 0 {
			deck = append(deck, side...)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	for i := 0; g.Turn.Step != game.StepDraw; i++ {
		if i >= 8 {
			t.Fatalf("never reached the draw step")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

func SideNamed(p *game.Player) []string {
	var out []string
	for _, c := range p.SideboardCards() {
		out = append(out, c.Name)
	}
	return out
}

func OutsideHandHas(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func OutsideExileHas(g *game.Game, name string) bool {
	for _, c := range g.Exile.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func OutsideGraveyardHas(p *game.Player, name string) bool {
	for _, c := range p.Graveyard.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func SideCardID(p *game.Player, name string) uuid.UUID {
	for _, c := range p.SideboardCards() {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func CandidateNames(g *game.Game, c *game.PendingChoice) []string {
	var out []string
	g.WithWriteLock(func() {
		for _, id := range c.ChooseCards {
			if card, ok := g.LookupOutsideCardForEffect(id); ok {
				out = append(out, card.Name)
			}
		}
	})
	return out
}

func EqualNames(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// cardKnownTo reports whether `viewer` is a knower of the sideboard card.
func CardKnownTo(p *game.Player, id, viewer uuid.UUID) bool {
	for _, c := range p.SideboardCards() {
		if c.InstanceID == id {
			return c.IsKnownTo(viewer)
		}
	}
	return false
}

func CardKnownToAll(g *game.Game, p *game.Player, id uuid.UUID) bool {
	for _, s := range g.Seats {
		if !CardKnownTo(p, id, s.ID) {
			return false
		}
	}
	return true
}
