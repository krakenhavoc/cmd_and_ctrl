package protocol

// citys_blessing_view_test.go: #2696, CR 702.131c. PlayerView carries the
// city's blessing, public and identical for every viewer, like the
// monarch. It is a designation on the table, not a secret.

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestPlayerViewCarriesTheCitysBlessingForEveryViewer(t *testing.T) {
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
	blessed := ids[1]
	g.Seats[1].CitysBlessing = true

	for label, viewer := range map[string]string{
		"the blessed seat": blessed,
		"another seat":     ids[0],
		"a spectator":      SpectatorViewerID,
		"the admin":        "",
	} {
		view := FilterViewFor(ViewOfGame(g), viewer)
		for _, s := range view.Seats {
			if want := s.ID == blessed; s.CitysBlessing != want {
				t.Errorf("%s: seat %s citys_blessing = %v, want %v", label, s.Name, s.CitysBlessing, want)
			}
		}
	}

	// On the wire it is omitted until earned, so a table that never saw
	// an ascend card pays no bytes for it.
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
		_, has := s["citys_blessing"]
		if want := s["id"] == blessed; has != want {
			t.Errorf("seat %v: citys_blessing on the wire = %v, want %v", s["name"], has, want)
		}
	}
}

// The log narrates the grant, once, naming the player.
func TestTheLogNarratesTheCitysBlessing(t *testing.T) {
	e := LogEvent{Kind: LogCitysBlessing}
	e.actorName = "Pat"
	if got, want := renderLogText(e, "", ""), "Pat gets the city's blessing"; got != want {
		t.Errorf("log text = %q, want %q", got, want)
	}
}
