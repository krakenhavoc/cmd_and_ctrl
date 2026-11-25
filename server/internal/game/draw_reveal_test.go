package game

import (
	"testing"

	"github.com/google/uuid"
)

// draw_reveal_test.go — #6238: "The chosen player reveals each card they
// draw" (Booby Trap). CR 701.20a: "To reveal a card, show that card to
// all players for a brief time."

const drawRevealOracle = "test-draw-reveal"

func stubDrawRevealCatalog(t *testing.T) {
	t.Helper()
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef {
		if key == drawRevealOracle {
			return &CardDef{RevealsChosenPlayersDraws: true}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogLookup = prev })
}

// drawRevealSource puts the revealing artifact under `owner`, with
// `chosen` its chosen player.
func drawRevealSourceFor(g *Game, owner *Player, chosen uuid.UUID) uuid.UUID {
	id := permanentFor(g, owner, "Test Booby Trap", "Artifact", "{6}")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].OracleID = drawRevealOracle
			g.Battlefield.Cards[i].ChosenPlayer = chosen
		}
	}
	return id
}

func drawRevealEvents(g *Game, card uuid.UUID) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventRevealCards && ev.CardID == card {
			out = append(out, ev)
		}
	}
	return out
}

// Every card the chosen player draws is revealed to every seat as it is
// drawn; the other players' draws are not.
func TestChosenPlayerRevealsEachCardTheyDraw(t *testing.T) {
	stubDrawRevealCatalog(t)
	g := newFourPlayerActiveGame(t)
	me, them, other := g.Seats[0], g.Seats[1], g.Seats[2]
	trap := drawRevealSourceFor(g, me, them.ID)

	first := libraryCard(them, "First")
	if err := g.DrawCard(them.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	second := libraryCard(them, "Second")
	if err := g.DrawCard(them.ID); err != nil {
		t.Fatalf("second draw: %v", err)
	}
	for _, id := range []uuid.UUID{first, second} {
		evs := drawRevealEvents(g, id)
		if len(evs) != 1 || evs[0].Source != trap || evs[0].Actor != them.ID {
			t.Fatalf("reveal events for a drawn card = %+v, want one from the trap", evs)
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.IsKnownTo(me.ID) || !c.IsKnownTo(other.ID) {
			t.Errorf("the drawn card is not known to every seat")
		}
	}

	mine := libraryCard(other, "Unrevealed")
	if err := g.DrawCard(other.ID); err != nil {
		t.Fatalf("other draw: %v", err)
	}
	if evs := drawRevealEvents(g, mine); len(evs) != 0 {
		t.Errorf("a player who is not chosen revealed a draw: %+v", evs)
	}
	if c, _ := g.LookupCardForEffect(mine); c.IsKnownTo(me.ID) {
		t.Error("an unchosen player's draw is known to another seat")
	}
}

// With the permanent gone, nothing is revealed.
func TestDrawRevealEndsWithItsSource(t *testing.T) {
	stubDrawRevealCatalog(t)
	g := newFourPlayerActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	trap := drawRevealSourceFor(g, me, them.ID)
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, trap); err != nil {
			t.Fatalf("remove the trap: %v", err)
		}
	})
	id := libraryCard(them, "Hidden")
	if err := g.DrawCard(them.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if evs := drawRevealEvents(g, id); len(evs) != 0 {
		t.Errorf("a draw was revealed with no source: %+v", evs)
	}
}
