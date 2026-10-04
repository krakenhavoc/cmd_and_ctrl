package protocol

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opening_roll_view_test.go — ADR 0121 §3: GameView.opening_roll and
// the opening roll's log lines.

// openingRollGame seats Alice, Bob, Carol and Dave and opens the roll on
// a key whose round 1 is 7, 4, 8, 8 and whose round 2 Dave wins 18 to 9.
func openingRollGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i, name := range []string{"Alice", "Bob", "Carol", "Dave"} {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(name, deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(22, 2048))); err != nil {
		t.Fatalf("StartWithOpeningRoll: %v", err)
	}
	return g
}

func TestOpeningRollView(t *testing.T) {
	g := openingRollGame(t)
	view := ViewOfGame(g)
	if view.OpeningRoll == nil || !reflect.DeepEqual(view.OpeningRoll.Rounds[0].Seats, []int{0, 1, 2, 3}) {
		t.Fatalf("opening_roll before any die = %+v", view.OpeningRoll)
	}
	raw, _ := json.Marshal(view.OpeningRoll)
	if want := `{"rounds":[{"seats":[0,1,2,3],"rolls":[]}]}`; string(raw) != want {
		t.Fatalf("wire = %s, want %s", raw, want)
	}

	if err := g.RollOpening(g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.HostRollRemaining(g.Seats[0].ID); err != nil {
		t.Fatal(err)
	}
	view = ViewOfGame(g)
	raw, _ = json.Marshal(view.OpeningRoll)
	want := `{"rounds":[{"seats":[0,1,2,3],"rolls":[{"seat":2,"result":8},{"seat":0,"result":7},{"seat":1,"result":4,"by":0},{"seat":3,"result":8,"by":0}]},{"seats":[2,3],"rolls":[]}]}`
	if string(raw) != want {
		t.Fatalf("wire =\n %s\nwant\n %s", raw, want)
	}
	// Public: every viewer, a spectator included, sees the same roll.
	for _, viewer := range []string{g.Seats[1].ID.String(), ""} {
		if got := FilterViewFor(view, viewer).OpeningRoll; !reflect.DeepEqual(got, view.OpeningRoll) {
			t.Fatalf("viewer %q sees %+v", viewer, got)
		}
	}

	if err := g.HostRollRemaining(uuid.Nil); err != nil {
		t.Fatal(err)
	}
	view = ViewOfGame(g)
	if c := view.OpeningRoll.Chooser; c == nil || *c != 3 {
		t.Fatalf("chooser = %v, want 3", c)
	}
	if by := view.OpeningRoll.Rounds[1].Rolls[0].By; by == nil || *by != game.OpeningRollByAdmin {
		t.Fatalf("the admin's die by = %v, want -1", by)
	}
	if err := g.ChooseStartingPlayer(g.Seats[3].ID, 3); err != nil {
		t.Fatal(err)
	}
	view = ViewOfGame(g)
	if view.OpeningRoll != nil || view.StartingSeat != 3 {
		t.Fatalf("after the choice: opening_roll %+v, starting_seat %d", view.OpeningRoll, view.StartingSeat)
	}

	// Every line, in order, as the table reads it.
	var lines []string
	for _, e := range view.Log {
		if e.Kind == LogRoll || e.Kind == LogOpeningRoll || e.Kind == LogStartingPlayer {
			lines = append(lines, e.Text)
		}
	}
	wantLines := []string{
		"Carol rolled a d20: 8",
		"Alice rolled for Bob and Dave",
		"Alice rolled a d20: 7",
		"Bob rolled a d20: 4",
		"Dave rolled a d20: 8",
		"Carol and Dave tied with 8 and roll again",
		"The admin rolled for Carol and Dave",
		"Carol rolled a d20: 9",
		"Dave rolled a d20: 18",
		"Dave won the opening roll with 18",
		"Dave chose to take the first turn",
	}
	if strings.Join(lines, "\n") != strings.Join(wantLines, "\n") {
		t.Fatalf("log =\n  %s\nwant\n  %s", strings.Join(lines, "\n  "), strings.Join(wantLines, "\n  "))
	}
	host := findLog(t, view.Log, LogOpeningRoll)
	if host.Label != game.OpeningRollRolledFor || host.Seat != 0 || !reflect.DeepEqual(host.Seats, []int{1, 3}) {
		t.Fatalf("the host's line = %+v", host)
	}
	chose := findLog(t, view.Log, LogStartingPlayer)
	if chose.Seat != 3 || chose.TargetSeat == nil || *chose.TargetSeat != 3 {
		t.Fatalf("starting_player = %+v", chose)
	}
}

func TestStartingPlayerLineNamesASeatGivenTheFirstTurn(t *testing.T) {
	g := openingRollGame(t)
	if err := g.HostRollRemaining(uuid.Nil); err != nil {
		t.Fatal(err)
	}
	if err := g.HostRollRemaining(uuid.Nil); err != nil {
		t.Fatal(err)
	}
	if err := g.ChooseStartingPlayer(g.Seats[3].ID, 1); err != nil {
		t.Fatal(err)
	}
	chose := findLog(t, ViewOfGame(g).Log, LogStartingPlayer)
	if chose.Text != "Dave chose Bob to take the first turn" || chose.TargetSeat == nil || *chose.TargetSeat != 1 {
		t.Fatalf("starting_player = %+v", chose)
	}
}
