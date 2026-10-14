package game

import (
	"testing"

	"github.com/google/uuid"
)

// public_names_test.go — #3609. An event names a token while it can
// still be found, so the log can say "Goblin token" after the token has
// ceased to exist; and never names anything hidden.

func publicToken(g *Game, owner *Player, name string) uuid.UUID {
	id := pushToken(g, owner, name, "Token Creature — Goblin")
	g.WithWriteLock(func() {
		c, _ := g.LookupCardPtrForEffect(id)
		for _, s := range g.Seats {
			c.AddKnower(s.ID)
		}
	})
	return id
}

func lastPublicNamesEvent(g *Game, kind EventKind) Event {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == kind {
			return g.Events[i]
		}
	}
	return Event{}
}

func TestAnEventCapturesTheNameOfAPublicToken(t *testing.T) {
	g := newActiveGame(t)
	id := publicToken(g, g.Seats[0], "Goblin Instigator")
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventCounterPlaced, CardID: id})
	})
	ev := lastPublicNamesEvent(g, EventCounterPlaced)
	if got := ev.PublicNames[id]; got != "Goblin Instigator" {
		t.Errorf("PublicNames[token] = %q, want its name", got)
	}
}

func TestTheCeaseToExistEventKeepsTheTokensName(t *testing.T) {
	g := newActiveGame(t)
	id := publicToken(g, g.Seats[0], "Goblin Instigator")
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	settle(g)
	nowhere(t, g, id)

	var ev Event
	for _, e := range g.Events {
		if e.Kind == EventZoneMove && e.CardID == id && e.NewZone == "" {
			ev = e
		}
	}
	if ev.Seq == 0 {
		t.Fatal("no ceased-to-exist move was emitted")
	}
	if got := ev.PublicNames[id]; got != "Goblin Instigator" {
		t.Errorf("the ceasing event names the token %q, want its name (it is in no zone by emit time)", got)
	}
}

// A token that is not known to every seat, a face-down one and a plain
// card are never captured: the fallback must stay fail-closed.
func TestHiddenObjectsAreNeverNamedOnAnEvent(t *testing.T) {
	g := newActiveGame(t)

	oneSeat := pushToken(g, g.Seats[0], "Secret Token", "Token Creature — Spy")
	g.WithWriteLock(func() {
		c, _ := g.LookupCardPtrForEffect(oneSeat)
		c.AddKnower(g.Seats[0].ID)
	})

	faceDown := publicToken(g, g.Seats[0], "Face Down Token")
	g.WithWriteLock(func() {
		c, _ := g.LookupCardPtrForEffect(faceDown)
		c.FaceDown = true
	})

	card := NewCard("Hidden Card", g.Seats[1].ID)
	g.WithWriteLock(func() {
		for _, s := range g.Seats {
			card.AddKnower(s.ID)
		}
		g.Seats[1].Hand.PushTop(card)
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: oneSeat, Target: faceDown, Source: card.InstanceID})
	})
	ev := lastPublicNamesEvent(g, EventZoneMove)
	if len(ev.PublicNames) != 0 {
		t.Errorf("PublicNames = %v, want none: nothing here is public", ev.PublicNames)
	}
}
