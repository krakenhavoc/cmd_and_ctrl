package protocol

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attractions_view_test.go — ADR 0207 §6: the Attraction deck is a
// public count of face-down cards, the junkyard is public, and an
// Attraction's lit numbers ride its card.

func attractionViewGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := 0; j < 20; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if i == 0 {
			for j := 0; j < 10; j++ {
				c := game.NewCard(fmt.Sprintf("Secret Attraction %d", j), uuid.Nil)
				c.TypeLine = "Artifact — Attraction"
				c.AstrotoriumBack = true
				c.AttractionLights = []int{2, 6}
				deck = append(deck, c)
			}
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), deck); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.StartWithSource(nil); err != nil {
		t.Fatal(err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

func TestAttractionDeckIsACountAndTheJunkyardIsPublic(t *testing.T) {
	g := attractionViewGame(t)
	me := g.Seats[0]
	var opened []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.OpenAttractionsThenForEffect(me.ID, uuid.Nil, 2, func(_ *game.Game, got []uuid.UUID) error {
			opened = got
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if err := g.DestroyPermanentForEffect(opened[0]); err != nil {
			t.Fatal(err)
		}
	})
	for _, viewer := range []string{me.ID.String(), g.Seats[1].ID.String(), SpectatorViewerID, ""} {
		v := ViewOfGameFor(g, viewer)
		seat := v.Seats[0]
		if seat.AttractionDeckCount != 8 {
			t.Errorf("viewer %q: attraction_deck_count %d, want 8", viewer, seat.AttractionDeckCount)
		}
		if len(seat.Junkyard) != 1 || seat.Junkyard[0].InstanceID != opened[0].String() {
			t.Errorf("viewer %q: junkyard %+v", viewer, seat.Junkyard)
		} else if !slices.Equal(seat.Junkyard[0].AttractionLights, []int{2, 6}) {
			t.Errorf("viewer %q: junkyard card lights %v", viewer, seat.Junkyard[0].AttractionLights)
		}
		var onBoard *CardView
		for i := range v.Battlefield.Cards {
			if v.Battlefield.Cards[i].InstanceID == opened[1].String() {
				onBoard = &v.Battlefield.Cards[i]
			}
		}
		if onBoard == nil || !slices.Equal(onBoard.AttractionLights, []int{2, 6}) {
			t.Errorf("viewer %q: the open Attraction's lights are not on the wire: %+v", viewer, onBoard)
		}
		// The face-down deck's cards are on nobody's wire.
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range me.AttractionDeckCards() {
			if strings.Contains(string(raw), c.InstanceID.String()) || strings.Contains(string(raw), c.Name) {
				t.Errorf("viewer %q: the wire carries the face-down %s", viewer, c.Name)
			}
		}
		if v.Seats[1].AttractionDeckCount != 0 || len(v.Seats[1].Junkyard) != 0 {
			t.Errorf("viewer %q: a seat with no Attractions shows some", viewer)
		}
	}
}
