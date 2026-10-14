package game

import (
	"testing"

	"github.com/google/uuid"
)

func TestLobbyDecksIsADeepCopyOfEachSeatsDeck(t *testing.T) {
	g := NewGame()
	cmdr := NewCommander("Cmdr", uuid.Nil)
	a, b := NewCard("A", uuid.Nil), NewCard("B", uuid.Nil)
	a.Keywords = []string{"flying"}
	p, err := g.AddPlayer("Alice", []Card{a, cmdr, b})
	if err != nil {
		t.Fatal(err)
	}
	q, err := g.AddPlayer("Bob", []Card{NewCard("C", uuid.Nil)})
	if err != nil {
		t.Fatal(err)
	}

	decks := g.LobbyDecks()
	if len(decks) != 2 || len(decks[p.ID]) != 3 || len(decks[q.ID]) != 1 {
		t.Fatalf("decks = %v", decks)
	}
	got := decks[p.ID]
	if got[0].Name != "Cmdr" || !got[0].IsCommander || got[1].Name != "A" || got[2].Name != "B" {
		t.Errorf("order = %s, %s, %s: want the commander, then the library as installed", got[0].Name, got[1].Name, got[2].Name)
	}
	seen := map[uuid.UUID]bool{}
	for _, c := range p.Library.Cards {
		seen[c.InstanceID] = true
	}
	for _, c := range got {
		if seen[c.InstanceID] || c.InstanceID == uuid.Nil {
			t.Errorf("%s reuses an instance id of the game it was copied from", c.Name)
		}
	}
	got[1].Keywords[0] = "changed"
	if p.Library.Cards[0].Keywords[0] != "flying" {
		t.Error("the copy shares its keyword slice with the game")
	}

	if err := g.Start(nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if g.LobbyDecks() != nil {
		t.Error("a started game has no lobby decks to copy")
	}
}
