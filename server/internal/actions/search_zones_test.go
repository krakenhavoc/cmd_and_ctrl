package actions

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// searchZonesGame opens "search your library and/or graveyard" for the
// first seat, with one match in each zone and the library known to the
// seat so a shuffle shows as un-knowing it.
func searchZonesGame(t *testing.T) (g *game.Game, p *game.Player, choiceID string, lib, gy uuid.UUID) {
	t.Helper()
	g = newGame(t)
	p = g.Seats[0]
	lib, gy = uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		p.Library.Cards = nil
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Other", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
		p.Library.PushTop(game.Card{InstanceID: lib, Name: "Wanted", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
		for i := range p.Library.Cards {
			p.Library.Cards[i].AddKnower(p.ID)
		}
		p.Graveyard.PushTop(game.Card{InstanceID: gy, Name: "Wanted", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
		if err := g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: p.ID, Pred: func(c game.Card) bool { return c.Name == "Wanted" },
			Dest: game.ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
		}); err != nil {
			t.Fatalf("SearchLibraryThenForEffect: %v", err)
		}
	})
	choice := pendingChoiceOfKind(g, game.PendingChoiceSearchLibrary)
	if choice == nil {
		t.Fatal("no search prompt queued")
	}
	return g, p, choice.ID.String(), lib, gy
}

func dispatchResolve(t *testing.T, g *game.Game, p *game.Player, payload map[string]any) error {
	t.Helper()
	a, err := Decode(string(TypeResolveChoice), p.ID.String(), params(t, payload))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	return Dispatch(g, a)
}

func libraryKnown(g *game.Game, p *game.Player) int {
	known := 0
	g.ReadSnapshot(func() {
		for _, c := range p.Library.Cards {
			if c.KnownBy[p.ID] {
				known++
			}
		}
	})
	return known
}

// TestDispatchResolveChoiceSearchZones — #4519 at the wire's front door.
// {search_zones: [...]} answers the zone question of an and/or search; the
// pick that follows is the ordinary {card_ids}. The library is shuffled
// only if "library" was among the zones.
func TestDispatchResolveChoiceSearchZones(t *testing.T) {
	for _, tc := range []struct {
		name         string
		zones        []string
		wantShuffled bool
	}{
		{"the graveyard alone", []string{"graveyard"}, false},
		{"the library alone", []string{"library"}, true},
		{"both", []string{"graveyard", "library"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, p, id, _, _ := searchZonesGame(t)
			if err := dispatchResolve(t, g, p, map[string]any{"choice_id": id, "search_zones": tc.zones}); err != nil {
				t.Fatalf("search_zones: %v", err)
			}
			if err := dispatchResolve(t, g, p, map[string]any{"choice_id": id, "card_ids": []string{}}); err != nil {
				t.Fatalf("card_ids: %v", err)
			}
			if shuffled := libraryKnown(g, p) == 0; shuffled != tc.wantShuffled {
				t.Errorf("library shuffled = %v, want %v", shuffled, tc.wantShuffled)
			}
		})
	}
}

// A pick before the question is answered, and a bad answer to it, are
// refused at the front door and leave the prompt open.
func TestDispatchRefusesAPickBeforeTheZonesAndABadZoneAnswer(t *testing.T) {
	g, p, id, lib, _ := searchZonesGame(t)
	if err := dispatchResolve(t, g, p, map[string]any{"choice_id": id, "card_ids": []string{lib.String()}}); err == nil {
		t.Error("a library pick before the zones were chosen was accepted")
	}
	for name, zones := range map[string][]string{
		"empty":   {},
		"unknown": {"exile"},
		"hand":    {"hand"},
	} {
		if err := dispatchResolve(t, g, p, map[string]any{"choice_id": id, "search_zones": zones}); err == nil {
			t.Errorf("%s: a bad zone answer was accepted", name)
		}
	}
	if pendingChoiceOfKind(g, game.PendingChoiceSearchLibrary) == nil {
		t.Fatal("a refused answer closed the prompt")
	}
	if got := libraryKnown(g, p); got != 2 {
		t.Errorf("the library was disturbed by refused answers: %d known", got)
	}
}

// search_zones on a prompt that is not a search is refused.
func TestDispatchRefusesSearchZonesOnAnotherKindOfPrompt(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		p.Library.Cards = nil
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Wanted", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
		if err := g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player: p.ID, Pred: func(c game.Card) bool { return c.Name == "Wanted" },
			Dest: game.ZoneHand, Limit: 1, Shuffle: true,
		}); err != nil {
			t.Fatalf("SearchLibraryThenForEffect: %v", err)
		}
	})
	choice := pendingChoiceOfKind(g, game.PendingChoiceSearchLibrary)
	if choice == nil {
		t.Fatal("no search prompt queued")
	}
	if err := dispatchResolve(t, g, p, map[string]any{"choice_id": choice.ID.String(), "search_zones": []string{"library"}}); err == nil {
		t.Error("a zone answer to a library-only search was accepted")
	}
}
