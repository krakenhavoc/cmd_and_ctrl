package protocol

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// companion_view_test.go — ADR 0172 §5: a revealed companion is public,
// its {3} row is its owner's, and the reveal is a log line.

func companionViewGame(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := game.NewGame()
	var id uuid.UUID
	for i := 0; i < 2; i++ {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := 0; j < 20; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if i == 0 {
			c := game.NewCard("Jegantha, the Wellspring", uuid.Nil)
			c.TypeLine = "Legendary Creature — Elemental Elk"
			c.ManaCost = "{4}{R/G}"
			c.IsCompanion = true
			id = c.InstanceID
			deck = append(deck, c)
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
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	return g, id
}

func TestCompanionIsPublicAndItsRowIsTheOwners(t *testing.T) {
	g, id := companionViewGame(t)
	owner, opp := g.Seats[0].ID.String(), g.Seats[1].ID.String()
	for _, tc := range []struct {
		viewer  string
		wantRow bool
	}{
		{owner, true},
		{opp, false},
		{SpectatorViewerID, false},
	} {
		v := ViewOfGameFor(g, tc.viewer)
		c := v.Seats[0].Companion
		if c == nil {
			t.Fatalf("viewer %s: no companion on seat 0", tc.viewer)
		}
		if c.InstanceID != id.String() || c.Name != "Jegantha, the Wellspring" || !c.IsCompanion {
			t.Errorf("viewer %s: companion %q (%s) companion=%v", tc.viewer, c.Name, c.InstanceID, c.IsCompanion)
		}
		hasRow := false
		for _, sa := range c.SpecialActions {
			if sa.Kind == string(game.SpecialActionCompanion) {
				hasRow = true
				if sa.Cost != "{3}" || !sa.Available {
					t.Errorf("row %+v, want {3} and available in the owner's main phase", sa)
				}
			}
		}
		if hasRow != tc.wantRow {
			t.Errorf("viewer %s: {3} row %v, want %v", tc.viewer, hasRow, tc.wantRow)
		}
		if v.Seats[1].Companion != nil {
			t.Errorf("viewer %s: seat 1 has a companion", tc.viewer)
		}
	}
}

func TestCompanionRevealIsALogLine(t *testing.T) {
	g, _ := companionViewGame(t)
	v := ViewOfGameFor(g, g.Seats[1].ID.String())
	for _, e := range v.Log {
		if e.Kind == LogCompanionRevealed {
			if !strings.Contains(e.Text, "revealed Jegantha, the Wellspring as their companion") {
				t.Errorf("line %q", e.Text)
			}
			return
		}
	}
	t.Error("no companion_revealed line in the log")
}
