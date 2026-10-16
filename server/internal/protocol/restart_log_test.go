package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// restart_log_test.go — #3856, CR 727: a restarted game's log begins
// again, with one `game_restarted` line naming the restarter and the
// card, which goes first. The old game's lines are gone from the view:
// every card in the new game is a new object, so they could only say
// "a card".
func TestGameRestartedStartsTheLogAgain(t *testing.T) {
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	restarter := g.Seats[len(g.Seats)-1]
	source := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: source, Name: "Karn Liberated", TypeLine: "Legendary Planeswalker — Karn",
			Owner: restarter.ID, Controller: restarter.ID,
		})
	})
	if before := ViewOfGame(g); len(before.Log) == 0 {
		t.Fatal("the old game wrote no log to lose")
	}
	g.WithWriteLock(func() {
		g.RestartTheGameForEffect(restarter.ID, source, nil)
	})
	g.SettleResolution()

	v := ViewOfGame(g)
	if len(v.Log) == 0 {
		t.Fatal("the new game has no log")
	}
	first := v.Log[0]
	if first.Kind != LogGameRestarted {
		t.Fatalf("the log begins with %s %q, want the restart line", first.Kind, first.Text)
	}
	if first.Seat != restarter.Seat {
		t.Errorf("seat %d, want the restarter's %d", first.Seat, restarter.Seat)
	}
	if !strings.Contains(first.Text, "restarted the game") || !strings.Contains(first.Text, "Karn Liberated") || !strings.Contains(first.Text, "goes first") {
		t.Errorf("text = %q", first.Text)
	}
	for _, e := range v.Log[1:] {
		if e.Kind == LogGameRestarted {
			t.Errorf("a second restart line: %+v", e)
		}
	}
	// Every viewer gets the line, with the card's name: the source is
	// an object of the old game and nobody can be denied its name.
	for _, p := range g.Seats {
		fv := ViewOfGameFor(g, p.ID.String())
		if len(fv.Log) == 0 || fv.Log[0].Kind != LogGameRestarted || !strings.Contains(fv.Log[0].Text, "Karn Liberated") {
			t.Errorf("viewer %s: %+v", p.Name, fv.Log)
		}
	}
}
