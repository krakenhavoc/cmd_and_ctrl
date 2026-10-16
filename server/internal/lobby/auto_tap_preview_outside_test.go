package lobby

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_outside_test.go — #3955. A Wish play is previewed for
// its cost like any other cast, so the preview reads the caller's OWN card
// outside the game, which the card index does not reach (CR 400.11c); a
// card in another seat's container is a 404, as an unknown card is.

func mintOutside(t *testing.T, g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.QueueOutsidePickForEffect(game.OutsidePickPrompt{Player: owner, Question: "Wish"}) {
			t.Fatal("no collection pick")
		}
	})
	var choiceID uuid.UUID
	g.ReadSnapshot(func() { choiceID = g.PendingChoices[len(g.PendingChoices)-1].ID })
	if err := g.ResolveOutsideCollection(choiceID, owner, []string{name}); err != nil {
		t.Fatalf("ResolveOutsideCollection: %v", err)
	}
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			if p.ID != owner {
				continue
			}
			for _, c := range p.SideboardCards() {
				if c.Name == name {
					id = c.InstanceID
				}
			}
		}
	})
	if id == uuid.Nil {
		t.Fatalf("%s was not minted outside the game", name)
	}
	return id
}

func TestAutoTapPreviewPricesTheCallersCardOutsideTheGame(t *testing.T) {
	prev := game.OutsideCollection
	game.OutsideCollection = lobbyCollection{cards: []game.Card{
		{Name: "Wishful Bear", TypeLine: "Creature — Bear", ManaCost: "{2}"},
		{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{3}"},
	}}
	t.Cleanup(func() { game.OutsideCollection = prev })

	f := newPreviewFixture(t)
	g, err := f.lobby.LookupGame(f.gameID)
	if err != nil {
		t.Fatal(err)
	}
	mine := mintOutside(t, g, f.alice, "Wishful Bear")
	theirs := mintOutside(t, g, f.bob, "Their Bear")
	f.mountains(2)

	body := f.preview(mine, 0)
	if !body.OK || body.Cost != "{2}" {
		t.Errorf("preview of my card outside the game: %+v, want ok at {2}", body)
	}

	resp := f.get("/games/" + f.gameID.String() + "/auto-tap-preview?card=" + theirs.String())
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("preview of another seat's card outside the game: status %d, want 404", resp.StatusCode)
	}
}
