package protocol

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_hand_view_test.go pins the per-viewer behaviour of the
// reveal_card sandbox verb (#671, ADR 0028's 2026-10-13 amendment) at
// the wire: who sees a revealed or shown hand card, what the owner
// learns about who else knows, and when the knowledge goes away.

func threeSeatGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 3 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 12 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(3, 4))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// secretInHand puts a distinctly named card in seat's hand, known to
// its owner only, and returns its ID.
func secretInHand(t *testing.T, g *game.Game, seat *game.Player) uuid.UUID {
	t.Helper()
	c := game.NewCard("Secret Tech", seat.ID)
	g.WithWriteLock(func() {
		seat.Hand.PushTop(c)
		seat.Hand.Cards[len(seat.Hand.Cards)-1].KnownBy = map[uuid.UUID]bool{seat.ID: true}
	})
	return c.InstanceID
}

func handCard(v GameView, owner *game.Player, id uuid.UUID) (CardView, bool) {
	for _, s := range v.Seats {
		if s.ID != owner.ID.String() {
			continue
		}
		for _, c := range s.Hand.Cards {
			if c.InstanceID == id.String() {
				return c, true
			}
		}
	}
	return CardView{}, false
}

func TestRevealHandCardToTableReachesEverySeat(t *testing.T) {
	g := threeSeatGame(t)
	owner := g.Seats[0]
	id := secretInHand(t, g, owner)

	if err := g.RevealHandCard(id, uuid.Nil); err != nil {
		t.Fatalf("RevealHandCard: %v", err)
	}
	var events int
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventRevealCards && ev.CardID == id && ev.RevealSeq != 0 {
				events++
			}
		}
	})
	if events != 1 {
		t.Fatalf("reveal events = %d, want 1", events)
	}
	for _, viewer := range g.Seats {
		v := ViewOfGameFor(g, viewer.ID.String())
		c, ok := handCard(v, owner, id)
		if !ok || c.Name != "Secret Tech" || !c.KnownByYou {
			t.Errorf("seat %s does not see the revealed card: ok=%v %+v", viewer.Name, ok, c)
		}
		if len(v.Reveals) != 1 {
			t.Errorf("seat %s: %d reveal frames, want 1", viewer.Name, len(v.Reveals))
		}
	}
}

func TestShowHandCardToOnePlayerIsPerViewer(t *testing.T) {
	g := threeSeatGame(t)
	owner, shown, other := g.Seats[0], g.Seats[1], g.Seats[2]
	id := secretInHand(t, g, owner)

	if err := g.RevealHandCard(id, shown.ID); err != nil {
		t.Fatalf("RevealHandCard: %v", err)
	}

	// The shown seat sees it face up in the owner's hand.
	c, ok := handCard(ViewOfGameFor(g, shown.ID.String()), owner, id)
	if !ok || c.Name != "Secret Tech" || !c.KnownByYou {
		t.Errorf("shown seat cannot see the card: ok=%v %+v", ok, c)
	}
	// Another opponent still gets no card for it at all.
	if c, ok := handCard(ViewOfGameFor(g, other.ID.String()), owner, id); ok {
		t.Errorf("a third seat sees the shown card: %+v", c)
	}
	// Nor does the spectator.
	if c, ok := handCard(ViewOfGameFor(g, SpectatorViewerID), owner, id); ok {
		t.Errorf("the spectator sees the shown card: %+v", c)
	}
	// No reveal frame, no banner for anyone.
	for _, viewer := range g.Seats {
		if v := ViewOfGameFor(g, viewer.ID.String()); len(v.Reveals) != 0 {
			t.Errorf("seat %s got a reveal frame for a private show", viewer.Name)
		}
	}
	// known_by_seats: the owner's own card only.
	oc, _ := handCard(ViewOfGameFor(g, owner.ID.String()), owner, id)
	if !slices.Equal(oc.KnownBySeats, []string{shown.ID.String()}) {
		t.Errorf("owner's known_by_seats = %v, want [%s]", oc.KnownBySeats, shown.ID)
	}
	sc, _ := handCard(ViewOfGameFor(g, shown.ID.String()), owner, id)
	if len(sc.KnownBySeats) != 0 {
		t.Errorf("known_by_seats leaked to another viewer: %v", sc.KnownBySeats)
	}
	// The log names the target and never the card, for every viewer.
	for _, viewer := range allViewers(g) {
		entries := revealLogEntries(ViewOfGameFor(g, viewer).Log)
		if len(entries) != 1 {
			t.Fatalf("viewer %q: %d reveal log entries, want 1", viewer, len(entries))
		}
		if want := "P1 revealed a card from their hand to P2"; entries[0].Text != want {
			t.Errorf("viewer %q: log %q, want %q", viewer, entries[0].Text, want)
		}
	}
}

func TestShowHandCardRefusals(t *testing.T) {
	g := threeSeatGame(t)
	owner, shown := g.Seats[0], g.Seats[1]
	id := secretInHand(t, g, owner)

	if err := g.RevealHandCard(id, owner.ID); err != game.ErrCannotShowTo {
		t.Errorf("show to owner: %v, want ErrCannotShowTo", err)
	}
	if err := g.RevealHandCard(id, uuid.New()); err != game.ErrCannotShowTo {
		t.Errorf("show to a non-seat: %v, want ErrCannotShowTo", err)
	}
	if err := g.Concede(shown.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if err := g.RevealHandCard(id, shown.ID); err != game.ErrCannotShowTo {
		t.Errorf("show to an eliminated seat: %v, want ErrCannotShowTo", err)
	}
	if err := g.RevealHandCard(uuid.New(), uuid.Nil); err != game.ErrCardNotFound {
		t.Errorf("unknown card: %v, want ErrCardNotFound", err)
	}
	var bf uuid.UUID
	g.WithWriteLock(func() {
		c := game.NewCard("On The Field", owner.ID)
		g.Battlefield.PushTop(c)
		bf = c.InstanceID
	})
	if err := g.RevealHandCard(bf, uuid.Nil); err != game.ErrNotInHand {
		t.Errorf("battlefield card: %v, want ErrNotInHand", err)
	}
}

// The knowledge ends when the card leaves the hand for a hidden zone.
func TestShownHandCardForgottenInLibrary(t *testing.T) {
	g := threeSeatGame(t)
	owner, shown := g.Seats[0], g.Seats[1]
	id := secretInHand(t, g, owner)
	if err := g.RevealHandCard(id, shown.ID); err != nil {
		t.Fatalf("RevealHandCard: %v", err)
	}
	ref := func(k game.ZoneKind) game.ZoneRef { return game.ZoneRef{Kind: k, Owner: owner.ID} }
	if err := g.MoveCardByID(ref(game.ZoneHand), ref(game.ZoneLibrary), id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	g.ReadSnapshot(func() {
		for _, c := range owner.Library.Cards {
			if c.InstanceID != id {
				continue
			}
			if len(c.KnownBy) != 0 {
				t.Errorf("a card put into the library keeps knowers: %v", c.KnownBy)
			}
			return
		}
		t.Errorf("card not found in library")
	})
}
