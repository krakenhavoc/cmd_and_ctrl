package protocol

// enduring_story_view_test.go: #4475, CR 702.195b. PlayerView carries the
// enduring story, public and identical for every viewer, like the city's
// blessing. It is a designation on the table, not a secret.

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestPlayerViewCarriesTheEnduringStoryForEveryViewer(t *testing.T) {
	g := game.NewGame()
	var ids []string
	for i := range 3 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		p, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck)
		if err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
		ids = append(ids, p.ID.String())
	}
	if err := g.Start(rand.New(rand.NewPCG(7, 8))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	storied := ids[1]
	g.Seats[1].EnduringStory = true
	g.Seats[2].CitysBlessing = true // the other designation leaves this one alone

	for label, viewer := range map[string]string{
		"the storied seat": storied,
		"another seat":     ids[0],
		"a spectator":      SpectatorViewerID,
		"the admin":        "",
	} {
		view := FilterViewFor(ViewOfGame(g), viewer)
		for _, s := range view.Seats {
			if want := s.ID == storied; s.EnduringStory != want {
				t.Errorf("%s: seat %s enduring_story = %v, want %v", label, s.Name, s.EnduringStory, want)
			}
		}
	}

	// On the wire it is omitted until earned, so a table that never saw
	// a storied card pays no bytes for it.
	raw, err := json.Marshal(ViewOfGame(g))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var wire struct {
		Seats []map[string]any `json:"seats"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, s := range wire.Seats {
		_, has := s["enduring_story"]
		if want := s["id"] == storied; has != want {
			t.Errorf("seat %v: enduring_story on the wire = %v, want %v", s["name"], has, want)
		}
	}
}

// The log narrates the grant, once, naming the player.
func TestTheLogNarratesTheEnduringStory(t *testing.T) {
	e := LogEvent{Kind: LogEnduringStory}
	e.actorName = "Pat"
	if got, want := renderLogText(e, "", ""), "Pat has an enduring story"; got != want {
		t.Errorf("log text = %q, want %q", got, want)
	}
}
