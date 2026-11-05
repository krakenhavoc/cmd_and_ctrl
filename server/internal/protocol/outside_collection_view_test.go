package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// outside_collection_view_test.go — ADR 0198 §3: a pick from the card
// database shows its suggestions, as cards, to its chooser alone.

type viewCollection struct{ cards []game.Card }

func (f viewCollection) Search(q string, limit int) []game.Card { return nil }
func (f viewCollection) Named(name string) (game.Card, bool) {
	for _, c := range f.cards {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}
	return game.Card{}, false
}
func (f viewCollection) Suggest(fn func(game.Card) bool) {
	for _, c := range f.cards {
		if !fn(c) {
			return
		}
	}
}

func TestACollectionPickShowsItsSuggestionsToTheChooserAlone(t *testing.T) {
	wish := game.NewCard("Database Sorcery", uuid.Nil)
	wish.TypeLine = "Sorcery"
	wish.ScryfallID = uuid.NewSHA1(uuid.NameSpaceURL, []byte("db")).String()
	bolt := game.NewCard("Database Instant", uuid.Nil)
	bolt.TypeLine = "Instant"
	prev := game.OutsideCollection
	game.OutsideCollection = viewCollection{cards: []game.Card{wish, bolt}}
	t.Cleanup(func() { game.OutsideCollection = prev })

	g, _ := sideboardViewGame(t)
	// Seat 1's deck has no sideboard: its pool is the database.
	chooser := g.Seats[1]
	g.WithWriteLock(func() {
		if !g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
			Player: chooser.ID, Question: "Burning Wish — reveal a sorcery card",
			Keyed: testProtocolSorceryPick,
		}) {
			t.Fatal("no prompt")
		}
	})
	own := ViewOfGameFor(g, chooser.ID.String())
	if len(own.PendingChoices) != 1 {
		t.Fatalf("%d pending choices", len(own.PendingChoices))
	}
	pc := own.PendingChoices[0]
	if pc.Kind != string(game.PendingChoiceOutsideCollection) || pc.ChooseMax != 1 || pc.ChooseMin != 0 ||
		pc.ChooseDestination != string(game.ChooseFromOutside) {
		t.Fatalf("chooser's prompt %q min %d max %d destination %q", pc.Kind, pc.ChooseMin, pc.ChooseMax, pc.ChooseDestination)
	}
	if len(pc.CollectionOptions) != 1 || pc.CollectionOptions[0].Name != "Database Sorcery" ||
		pc.CollectionOptions[0].InstanceID != "" || pc.CollectionOptions[0].ScryfallID != wish.ScryfallID {
		t.Fatalf("chooser's suggestions %+v", pc.CollectionOptions)
	}
	for _, viewer := range []string{g.Seats[0].ID.String(), SpectatorViewerID, ""} {
		v := ViewOfGameFor(g, viewer)
		if len(v.PendingChoices) != 1 {
			t.Fatalf("viewer %q: %d pending choices", viewer, len(v.PendingChoices))
		}
		other := v.PendingChoices[0]
		if len(other.CollectionOptions) != 0 || other.ChooseMax != 0 || other.ChooseDestination != "" {
			t.Errorf("viewer %q sees %d suggestions, max %d", viewer, len(other.CollectionOptions), other.ChooseMax)
		}
		raw, _ := json.Marshal(v)
		if strings.Contains(string(raw), "Database Sorcery") {
			t.Errorf("viewer %q: the wire carries the chooser's suggestions", viewer)
		}
	}
}

// #4947: a database pick "with different names" says so, so the picker
// takes each name once, and a pick whose names may repeat (Research's
// "up to four cards") does not. Like the bounds, it is the chooser's.
func TestACollectionPickSaysWhenItsNamesMustDiffer(t *testing.T) {
	legend := game.NewCard("Database Legend", uuid.Nil)
	legend.TypeLine = "Legendary Creature — Turtle"
	prev := game.OutsideCollection
	game.OutsideCollection = viewCollection{cards: []game.Card{legend}}
	t.Cleanup(func() { game.OutsideCollection = prev })

	for _, tc := range []struct {
		name string
		pick game.OutsidePick
		want bool
	}{
		{"different names", testProtocolDistinctPick, true},
		{"names may repeat", testProtocolRepeatPick, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _ := sideboardViewGame(t)
			// Seat 1's deck has no sideboard: its pool is the database.
			chooser := g.Seats[1]
			g.WithWriteLock(func() {
				if !g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
					Player: chooser.ID, Question: "choose up to four cards", Max: 4,
					Keyed: tc.pick,
				}) {
					t.Fatal("no prompt")
				}
			})
			own := ViewOfGameFor(g, chooser.ID.String())
			if len(own.PendingChoices) != 1 {
				t.Fatalf("%d pending choices", len(own.PendingChoices))
			}
			pc := own.PendingChoices[0]
			if pc.Kind != string(game.PendingChoiceOutsideCollection) || pc.ChooseMax != 4 {
				t.Fatalf("chooser's prompt %q max %d", pc.Kind, pc.ChooseMax)
			}
			if pc.DistinctNames != tc.want {
				t.Fatalf("distinct_names = %v, want %v", pc.DistinctNames, tc.want)
			}
			raw, _ := json.Marshal(pc)
			if got := strings.Contains(string(raw), `"distinct_names":true`); got != tc.want {
				t.Fatalf("the chooser's wire carries distinct_names: %v, want %v (%s)", got, tc.want, raw)
			}
			for _, viewer := range []string{g.Seats[0].ID.String(), SpectatorViewerID, ""} {
				v := ViewOfGameFor(g, viewer)
				if len(v.PendingChoices) != 1 {
					t.Fatalf("viewer %q: %d pending choices", viewer, len(v.PendingChoices))
				}
				if v.PendingChoices[0].DistinctNames {
					t.Errorf("viewer %q sees the chooser's distinct_names", viewer)
				}
				raw, _ := json.Marshal(v)
				if strings.Contains(string(raw), "distinct_names") {
					t.Errorf("viewer %q: the wire carries distinct_names", viewer)
				}
			}
		})
	}
}
