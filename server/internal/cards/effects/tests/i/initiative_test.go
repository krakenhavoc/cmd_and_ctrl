package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestTakingTheInitiativeVenturesIntoSecretEntrance is the headline: a
// card's "you take the initiative" hands over the designation, and the
// inherent "whenever a player takes the initiative" trigger enters the
// Undercity's first room, whose ability searches for a basic land.
func TestTakingTheInitiativeVenturesIntoSecretEntrance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	effectstest.InitTake(t, g, me.ID)
	if g.Initiative != me.ID {
		t.Fatalf("initiative = %v, want %v", g.Initiative, me.ID)
	}
	if room := effectstest.InitRoom(g, me.ID); room != "" {
		t.Fatalf("ventured before the trigger resolved: %q", room)
	}
	effectstest.MonSettle(t, g)
	if room := effectstest.InitRoom(g, me.ID); room != "secret-entrance" {
		t.Fatalf("room = %q, want secret-entrance", room)
	}
	c := effectstest.InitPending(g, game.PendingChoiceSearchLibrary, me.ID)
	if c == nil {
		t.Fatalf("Secret Entrance asked for no search (pending %d)", len(g.PendingChoices))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if z := g.FindCardZoneForEffect(forest); z == nil || z.Kind != game.ZoneHand {
		t.Fatalf("the Forest is not in hand")
	}
}
