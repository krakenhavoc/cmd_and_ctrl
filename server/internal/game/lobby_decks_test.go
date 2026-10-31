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

// CommandersOf (ADR 0212 phase B2) is how the lobby reads a seat's
// commanders again from a restored game: the command zone in the lobby,
// in the deck's order, and every zone once the game is under way, since
// the designation follows the card (CR 903.3). A token copy is not one
// (CR 707.2: the designation is no characteristic for a copy to take).
func TestCommandersOfFollowsTheCardAndSkipsTokens(t *testing.T) {
	g := NewGame()
	tymna, kraum := NewCommander("Tymna the Weaver", uuid.Nil), NewCommander("Kraum, Ludevic's Opus", uuid.Nil)
	p, err := g.AddPlayer("Alice", []Card{tymna, NewCard("A", uuid.Nil), kraum})
	if err != nil {
		t.Fatal(err)
	}
	q, err := g.AddPlayer("Bob", []Card{NewCommander("Bob's Commander", uuid.Nil), NewCard("B", uuid.Nil)})
	if err != nil {
		t.Fatal(err)
	}
	names := func(cs []Card) []string {
		var out []string
		for _, c := range cs {
			out = append(out, c.Name)
		}
		return out
	}

	got := g.CommandersOf(p.ID)
	if len(got) != 2 || got[0].Name != "Tymna the Weaver" || got[1].Name != "Kraum, Ludevic's Opus" {
		t.Fatalf("in the lobby: %v, want the partner pair in deck order", names(got))
	}
	got[0].Name = "changed"
	if p.Command.Cards[0].Name != "Tymna the Weaver" {
		t.Error("CommandersOf hands out the game's own card")
	}
	if g.CommandersOf(uuid.New()) != nil {
		t.Error("a player who is not seated has commanders")
	}

	if err := g.Start(nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	cast := p.Command.Cards[0].InstanceID
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneCommand, Owner: p.ID}, ZoneRef{Kind: ZoneBattlefield}, cast); err != nil {
		t.Fatalf("move to the battlefield: %v", err)
	}
	g.WithWriteLock(func() {
		tok := cloneCard(p.Command.Cards[0])
		tok.InstanceID = uuid.New()
		tok.TypeLine = "Token Legendary Creature — Human Cleric"
		g.Battlefield.PushTop(tok)
	})
	got = g.CommandersOf(p.ID)
	if len(got) != 2 {
		t.Fatalf("under way: %v, want both commanders and not the token", names(got))
	}
	seen := map[string]bool{}
	for _, c := range got {
		seen[c.Name] = true
		if c.IsToken() {
			t.Errorf("%s is a token", c.Name)
		}
	}
	if !seen["Tymna the Weaver"] || !seen["Kraum, Ludevic's Opus"] {
		t.Errorf("under way: %v, want Tymna (on the battlefield) and Kraum (in the command zone)", names(got))
	}
	if b := g.CommandersOf(q.ID); len(b) != 1 || b[0].Name != "Bob's Commander" {
		t.Errorf("Bob's commanders = %v", names(b))
	}
}
