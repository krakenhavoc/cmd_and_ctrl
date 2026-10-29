package actions

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestDispatchResolveChoiceSearchLibrarySearched — #4516 at the wire's
// front door. An empty resolve_choice for a "search your library and/or
// graveyard" prompt is a search of the graveyard alone: the library keeps
// its order and no search is announced. {library_searched: true} says the
// library was looked in too, and it shuffles.
func TestDispatchResolveChoiceSearchLibrarySearched(t *testing.T) {
	for _, tc := range []struct {
		name         string
		payload      map[string]any
		wantShuffled bool
	}{
		{"absent is the graveyard alone", map[string]any{}, false},
		{"false is the graveyard alone", map[string]any{"library_searched": false}, false},
		{"true is a library search", map[string]any{"library_searched": true}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newGame(t)
			p := g.Seats[0]
			g.WithWriteLock(func() {
				p.Library.Cards = nil
				for _, n := range []string{"Other", "Another"} {
					p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Land", Owner: p.ID, Controller: p.ID})
				}
				for i := range p.Library.Cards {
					p.Library.Cards[i].AddKnower(p.ID)
				}
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
			payload := map[string]any{"choice_id": choice.ID.String(), "card_ids": []string{}}
			for k, v := range tc.payload {
				payload[k] = v
			}
			a, err := Decode(string(TypeResolveChoice), p.ID.String(), params(t, payload))
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if err := Dispatch(g, a); err != nil {
				t.Fatalf("Dispatch: %v", err)
			}
			known := 0
			g.ReadSnapshot(func() {
				for _, c := range p.Library.Cards {
					if c.KnownBy[p.ID] {
						known++
					}
				}
			})
			if shuffled := known == 0; shuffled != tc.wantShuffled {
				t.Errorf("library shuffled = %v, want %v", shuffled, tc.wantShuffled)
			}
		})
	}
}
