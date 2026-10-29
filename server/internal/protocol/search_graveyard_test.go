package protocol

import (
	"encoding/json"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// searchZonesTable seats two players with a Command Tower in the first
// one's library and graveyard, opens "search your library and/or
// graveyard" for it, and returns the table and the two ids.
func searchZonesTable(t *testing.T) (g *game.Game, searcher, opponent *game.Player, libID, gyID uuid.UUID) {
	t.Helper()
	g = game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 10)
		for j := range deck {
			deck[j] = game.NewCard("filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer %d: %v", i, err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	searcher, opponent = g.Seats[0], g.Seats[1]
	searcher.Library.Cards = nil
	libID, gyID = uuid.New(), uuid.New()
	searcher.Library.PushTop(game.Card{InstanceID: libID, Name: "Command Tower", TypeLine: "Land", Owner: searcher.ID, Controller: searcher.ID})
	searcher.Graveyard.PushTop(game.Card{InstanceID: gyID, Name: "Command Tower", TypeLine: "Land", Owner: searcher.ID, Controller: searcher.ID})
	g.WithWriteLock(func() {
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: searcher.ID, Pred: func(c game.Card) bool { return c.Name == "Command Tower" },
			Dest: game.ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
		})
	})
	return g, searcher, opponent, libID, gyID
}

// searchChoiceID is the one open choice's id in a view.
func searchChoiceID(t *testing.T, v GameView) uuid.UUID {
	t.Helper()
	if len(v.PendingChoices) != 1 {
		t.Fatalf("%d choices, want 1", len(v.PendingChoices))
	}
	return uuid.MustParse(v.PendingChoices[0].ID)
}

// #4519: a "search your library and/or graveyard" prompt asks which zones
// first. Until it is answered the searcher is shown no card (looking
// through the library is searching it), no floor and no zone list; they
// are told it is the zone question and which zone is on offer beside the
// library. The opponent is told none of it.
func TestSearchPromptAsksTheZonesBeforeShowingAnything(t *testing.T) {
	g, searcher, opponent, libID, _ := searchZonesTable(t)

	mine := FilterViewFor(ViewOfGame(g), searcher.ID.String())
	if len(mine.PendingChoices) != 1 {
		t.Fatalf("searcher sees %d choices, want 1", len(mine.PendingChoices))
	}
	pc := mine.PendingChoices[0]
	if !pc.SearchZonesPending || pc.SearchOtherZone != "graveyard" {
		t.Fatalf("search_zones_pending/search_other_zone = %v/%q, want true/graveyard", pc.SearchZonesPending, pc.SearchOtherZone)
	}
	if len(pc.Options) != 0 || len(pc.SearchGraveyard) != 0 || pc.SearchMin != 0 {
		t.Errorf("the zone question shows candidates: %+v", pc)
	}
	raw, _ := json.Marshal(mine.PendingChoices)
	if strings.Contains(string(raw), libID.String()) {
		t.Error("the library card's id is on the wire before the library was chosen")
	}

	theirs := FilterViewFor(ViewOfGame(g), opponent.ID.String())
	if their := theirs.PendingChoices[0]; their.SearchZonesPending || their.SearchOtherZone != "" ||
		len(their.SearchGraveyard) != 0 || len(their.Options) != 0 {
		t.Errorf("the opponent sees the search: %+v", their)
	}
}

// #3217, #4519: with both zones chosen the prompt offers both Towers,
// marks the graveyard's (search_graveyard), and leaves search_min absent:
// "and/or" lets the searcher find nothing. The opponent's view of the
// prompt is the same before and after the answer.
func TestSearchPromptMarksTheGraveyardHalf(t *testing.T) {
	g, searcher, opponent, _, gyID := searchZonesTable(t)
	before := FilterViewFor(ViewOfGame(g), opponent.ID.String())
	id := searchChoiceID(t, before)
	if err := g.ChooseSearchZones(id, searcher.ID, []string{game.SearchZoneLibrary, game.SearchZoneGraveyard}); err != nil {
		t.Fatal(err)
	}

	pc := FilterViewFor(ViewOfGame(g), searcher.ID.String()).PendingChoices[0]
	if pc.SearchZonesPending || pc.SearchLibrarySkipped || pc.SearchOtherZone != "graveyard" {
		t.Errorf("an answered question is still advertised, or the zone is lost: %+v", pc)
	}
	if len(pc.Options) != 2 {
		t.Fatalf("options = %d, want both Towers", len(pc.Options))
	}
	if len(pc.SearchGraveyard) != 1 || pc.SearchGraveyard[0] != gyID.String() {
		t.Errorf("search_graveyard = %v, want just the graveyard Tower %s", pc.SearchGraveyard, gyID)
	}
	if pc.SearchMin != 0 {
		t.Errorf("search_min = %d, want 0: finding nothing stays legal", pc.SearchMin)
	}

	after := FilterViewFor(ViewOfGame(g), opponent.ID.String())
	b, _ := json.Marshal(before.PendingChoices)
	a, _ := json.Marshal(after.PendingChoices)
	if string(a) != string(b) {
		t.Errorf("the opponent's view changed with the searcher's zone answer:\n before %s\n after  %s", b, a)
	}
}

// #4519: graveyard only. The library's Tower is in nobody's view, and the
// graveyard's is the only option.
func TestSearchPromptGraveyardOnlyNeverShowsTheLibrary(t *testing.T) {
	g, searcher, opponent, libID, gyID := searchZonesTable(t)
	id := searchChoiceID(t, FilterViewFor(ViewOfGame(g), searcher.ID.String()))
	if err := g.ChooseSearchZones(id, searcher.ID, []string{game.SearchZoneGraveyard}); err != nil {
		t.Fatal(err)
	}
	for _, viewer := range []string{searcher.ID.String(), opponent.ID.String()} {
		raw, _ := json.Marshal(FilterViewFor(ViewOfGame(g), viewer).PendingChoices)
		if strings.Contains(string(raw), libID.String()) {
			t.Errorf("viewer %s sees the library card of a graveyard-only search", viewer)
		}
	}
	pc := FilterViewFor(ViewOfGame(g), searcher.ID.String()).PendingChoices[0]
	if len(pc.Options) != 1 || pc.Options[0].InstanceID != gyID.String() {
		t.Errorf("options = %+v, want only the graveyard Tower", pc.Options)
	}
	if !pc.SearchLibrarySkipped {
		t.Error("search_library_skipped is absent from a graveyard-only search")
	}
	if their := FilterViewFor(ViewOfGame(g), opponent.ID.String()).PendingChoices[0]; their.SearchLibrarySkipped {
		t.Error("the opponent is told the library was skipped")
	}
}
