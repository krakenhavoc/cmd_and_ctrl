package effectstest

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// openingTable starts a four-seat game and stops with the mulligan
// window open, so a test can put exact cards into the opening hands
// before anyone keeps. Seat 0 is the starting player.
func OpeningTable(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g
}

// handWith replaces a seat's opening hand with exactly these cards and
// returns their IDs in order.
func HandWith(g *game.Game, seat int, cards ...game.Card) []uuid.UUID {
	p := g.Seats[seat]
	p.Hand.Cards = nil
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		c.KnownBy = map[uuid.UUID]bool{p.ID: true}
		p.Hand.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func LeylineOfSanctity() game.Card {
	return game.Card{Name: "Leyline of Sanctity", TypeLine: "Enchantment", ManaCost: "{2}{W}{W}", OracleID: LeylineOfSanctityOracle}
}

func GemstoneCaverns() game.Card {
	return game.Card{Name: "Gemstone Caverns", TypeLine: "Legendary Land", OracleID: GemstoneCavernsOracle}
}

func FillerCard(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Sorcery", ManaCost: "{1}{R}"}
}

// keepAll keeps every seat's hand in turn order, closing the window.
func KeepAll(t *testing.T, g *game.Game) {
	t.Helper()
	for i := range g.Seats {
		if err := g.KeepHand(g.Seats[(g.StartingSeat+i)%len(g.Seats)].ID); err != nil {
			t.Fatalf("KeepHand seat %d: %v", i, err)
		}
	}
}

func ChoicesOf(g *game.Game, kind game.PendingChoiceKind) []*game.PendingChoice {
	var out []*game.PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c.Kind == kind {
				out = append(out, c)
			}
		}
	})
	return out
}

func BeganOnBattlefield(g *game.Game, id uuid.UUID) bool {
	var ok bool
	g.ReadSnapshot(func() { ok = g.Battlefield.Contains(id) })
	return ok
}
