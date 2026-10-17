package effects

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// initiative_view_test.go pins what the Undercity shows to whom (ADR
// 0203): the dungeon card and its marker are public; Lost Well's scry
// is seen only by the scrying player; the Throne's reveal reaches the
// table.

func initViewJSON(t *testing.T, g *game.Game, viewer uuid.UUID) string {
	t.Helper()
	raw, err := json.Marshal(protocol.FilterViewFor(protocol.ViewOfGame(g), viewer.String()))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(raw)
}

// TestTheDungeonIsPublic: every seat sees the same dungeon, room and
// completed list for a player.
func TestTheDungeonIsPublic(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	initCompleteADungeon(me)
	initVenture(t, g, me.ID)
	for _, viewer := range []uuid.UUID{me.ID, them.ID} {
		v := protocol.FilterViewFor(protocol.ViewOfGame(g), viewer.String())
		var seat *protocol.PlayerView
		for i := range v.Seats {
			if v.Seats[i].ID == me.ID.String() {
				seat = &v.Seats[i]
			}
		}
		if seat == nil || seat.Dungeon == nil {
			t.Fatalf("viewer %v: no dungeon on the seat", viewer)
		}
		if seat.Dungeon.Name != "Undercity" || seat.Dungeon.Room != "secret-entrance" || len(seat.Dungeon.Rooms) != 9 {
			t.Errorf("viewer %v: dungeon %+v", viewer, seat.Dungeon)
		}
		if len(seat.CompletedDungeons) != 1 || seat.CompletedDungeons[0] != "Undercity" {
			t.Errorf("viewer %v: completed %v", viewer, seat.CompletedDungeons)
		}
		if v.Initiative != me.ID.String() {
			t.Errorf("viewer %v: initiative %q", viewer, v.Initiative)
		}
	}
}

// TestLostWellScryIsSeenOnlyByTheScryingPlayer.
func TestLostWellScryIsSeenOnlyByTheScryingPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	initVenture(t, g, me.ID)
	for _, name := range []string{"Hidden Alpha", "Hidden Beta"} {
		c := game.Card{Name: name, TypeLine: "Instant", InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID}
		me.Library.PushTop(c)
	}
	initTake(t, g, me.ID)
	for i := 0; i < 16 && initPending(g, game.PendingChoiceScry, me.ID) == nil; i++ {
		if c := initPending(g, game.PendingChoiceOptionPick, me.ID); c != nil {
			initPickRoom(t, g, me.ID, "lost-well")
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if initPending(g, game.PendingChoiceScry, me.ID) == nil {
		t.Fatalf("Lost Well asked for no scry")
	}
	mine, theirs := initViewJSON(t, g, me.ID), initViewJSON(t, g, them.ID)
	for _, name := range []string{"Hidden Alpha", "Hidden Beta"} {
		if !strings.Contains(mine, name) {
			t.Errorf("the scrying player cannot see %q", name)
		}
		if strings.Contains(theirs, name) {
			t.Errorf("an opponent can see %q", name)
		}
	}
}

// TestTheThroneRevealReachesTheTable: the top ten are revealed to every
// seat.
func TestTheThroneRevealReachesTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	marker := game.Card{Name: "Revealed Giant", TypeLine: "Creature — Giant", Power: 3, Toughness: 3, InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(marker)
	initVenture(t, g, me.ID)
	initVenture(t, g, me.ID, "lost-well")
	initVenture(t, g, me.ID, "stash")
	initVenture(t, g, me.ID)
	// The marker card is still on top for the Throne: the scry kept the
	// order and nothing drew.
	initVenture(t, g, me.ID)
	if !containsName(revealedNames(revealsFor(g, them.ID)), "Revealed Giant") {
		t.Errorf("the Throne's reveal did not reach an opponent")
	}
}
