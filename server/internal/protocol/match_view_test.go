package protocol

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// match_view_test.go: GameView.match (ADR 0222 §3).

func matchDeck(owner string) []game.Card {
	deck := make([]game.Card, 0, 75)
	for i := 0; i < 60; i++ {
		deck = append(deck, game.NewCard(fmt.Sprintf("%s Main %d", owner, i+1), uuid.Nil))
	}
	for i := 0; i < 15; i++ {
		c := game.NewCard(fmt.Sprintf("%s Side %d", owner, i+1), uuid.Nil)
		c.IsSideboard = true
		deck = append(deck, c)
	}
	return deck
}

// standardTable is a started Standard table of the given match length,
// Alice in seat 0 and Bob in seat 1.
func standardTable(t *testing.T, bestOf int) *game.Game {
	t.Helper()
	g := game.NewGame()
	std := game.FormatStandard
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{Format: &std, BestOf: &bestOf}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Alice", "Bob"} {
		if _, err := g.AddPlayer(name, matchDeck(name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(3, 5))); err != nil {
		t.Fatal(err)
	}
	return g
}

// The score is on the view from the first game, and a table between
// games is in play on the wire with its sideboarding window open.
func TestMatchViewScoreAndSideboarding(t *testing.T) {
	g := standardTable(t, 3)
	alice, bob := g.Seats[0].ID, g.Seats[1].ID

	v := ViewOfGame(g)
	if v.Match == nil || v.Match.BestOf != 3 || v.Match.Game != 1 || v.Match.Sideboarding != nil {
		t.Fatalf("game 1's match view = %+v", v.Match)
	}
	if got := fmt.Sprint(v.Match.Wins); got != "[0 0]" {
		t.Errorf("wins = %s, want [0 0]", got)
	}
	if v.Settings == nil || v.Settings.BestOf != 3 {
		t.Errorf("settings.best_of = %+v, want 3", v.Settings)
	}

	if err := g.Concede(bob); err != nil {
		t.Fatal(err)
	}
	v = ViewOfGame(g)
	if v.State != "active" {
		t.Errorf("state on the wire between games = %q, want active", v.State)
	}
	if v.Outcome != nil {
		t.Errorf("a table between games has an outcome on the wire: %+v", v.Outcome)
	}
	m := v.Match
	if m == nil || m.Game != 1 || m.Over || fmt.Sprint(m.Wins) != "[1 0]" || len(m.Games) != 1 {
		t.Fatalf("match view between games = %+v", m)
	}
	if w := m.Games[0].WinnerSeat; w == nil || *w != 0 || m.Games[0].Cause != game.OutcomeCauseLastStanding {
		t.Errorf("game 1 on the wire = %+v", m.Games[0])
	}
	sb := m.Sideboarding
	if sb == nil {
		t.Fatal("no sideboarding window on the wire")
	}
	if sb.ChooserSeat != 1 || sb.PlayFirst == nil || !*sb.PlayFirst || sb.DeckMin != 60 || sb.SideboardMax != 15 {
		t.Errorf("sideboarding = %+v", sb)
	}
	if sb.SecondsLeft <= 0 || sb.SecondsLeft > int(game.SideboardingTime.Seconds()) || sb.Deadline == "" {
		t.Errorf("deadline %q, %d seconds left", sb.Deadline, sb.SecondsLeft)
	}
	if len(sb.Ready) != 0 || len(sb.Pools) != 2 {
		t.Fatalf("ready %v, %d pools; want nobody ready and both pools on the full view", sb.Ready, len(sb.Pools))
	}
	for _, p := range sb.Pools {
		if len(p.Deck) != 60 || len(p.Sideboard) != 15 || len(p.Registered) != 60 {
			t.Errorf("seat %d's pool is %d, %d and %d registered", p.Seat, len(p.Deck), len(p.Sideboard), len(p.Registered))
		}
	}
	for _, seat := range v.Seats {
		if f := FilterViewFor(v, seat.ID); len(f.LegalMoves) != 0 {
			t.Errorf("%s has %d legal moves on a finished board", seat.Name, len(f.LegalMoves))
		}
	}

	// Each player's frame carries their own pool and nobody else's;
	// the choice of who plays first is the chooser's alone.
	for _, tc := range []struct {
		name      string
		viewer    string
		pools     int
		playFirst bool
	}{
		{"the winner", alice.String(), 1, false},
		{"the loser, who chooses", bob.String(), 1, true},
		{"a spectator", SpectatorViewerID, 0, false},
		{"the admin", "", 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := FilterViewFor(v, tc.viewer)
			fsb := f.Match.Sideboarding
			if fsb == nil || fmt.Sprint(f.Match.Wins) != "[1 0]" {
				t.Fatalf("filtered match = %+v", f.Match)
			}
			if len(fsb.Pools) != tc.pools {
				t.Fatalf("%d pools, want %d", len(fsb.Pools), tc.pools)
			}
			if tc.pools == 1 && fsb.Pools[0].PlayerID != tc.viewer {
				t.Errorf("the pool on this frame is %s's", fsb.Pools[0].PlayerID)
			}
			if (fsb.PlayFirst != nil) != tc.playFirst {
				t.Errorf("play_first present = %v, want %v", fsb.PlayFirst != nil, tc.playFirst)
			}
			raw, err := json.Marshal(f)
			if err != nil {
				t.Fatal(err)
			}
			other := "Bob Side 1"
			if tc.viewer == bob.String() {
				other = "Alice Side 1"
			}
			if tc.pools < 2 && strings.Contains(string(raw), other) {
				t.Errorf("this frame names %q, a card of the other player's sideboard", other)
			}
		})
	}

	// Ready seats, then the next game: the log begins again with its line.
	if err := g.ChoosePlayFirst(bob, false); err != nil {
		t.Fatal(err)
	}
	if err := g.SetSideboardReady(alice, true); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprint(ViewOfGame(g).Match.Sideboarding.Ready); got != "[0]" {
		t.Errorf("ready = %s, want [0]", got)
	}
	if err := g.SetSideboardReady(bob, true); err != nil {
		t.Fatal(err)
	}
	v = ViewOfGame(g)
	if v.State != "active" || !v.MulligansOpen || v.Match.Game != 2 || v.Match.Sideboarding != nil || v.Outcome != nil {
		t.Fatalf("game 2's view: state %q, match %+v", v.State, v.Match)
	}
	if len(v.Log) == 0 || v.Log[0].Kind != LogMatchGame {
		t.Fatalf("the log does not begin with the game's line: %+v", v.Log)
	}
	if got, want := v.Log[0].Text, "Game 2. Bob chose to draw, so Alice plays first"; got != want {
		t.Errorf("log line = %q, want %q", got, want)
	}
	for _, e := range v.Log {
		if e.Kind == LogGameOver {
			t.Error("game 1's game-over line is in game 2's log")
		}
	}

	// The match ends: the table is ended, and the match names its winner.
	if err := g.Concede(bob); err != nil {
		t.Fatal(err)
	}
	v = ViewOfGame(g)
	if v.State != "ended" || v.Outcome == nil || !v.Match.Over || v.Match.WinnerSeat == nil || *v.Match.WinnerSeat != 0 {
		t.Fatalf("after 2-0: state %q, outcome %+v, match %+v", v.State, v.Outcome, v.Match)
	}
	if got := fmt.Sprint(v.Match.Wins); got != "[2 0]" {
		t.Errorf("wins = %s, want [2 0]", got)
	}
}

// A Commander table's view has no match and no match length, and its
// JSON names neither.
func TestCommanderViewHasNoMatch(t *testing.T) {
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := 0; j < 30; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	check := func(when string) {
		v := ViewOfGame(g)
		if v.Match != nil {
			t.Fatalf("%s: a Commander table's view has a match: %+v", when, v.Match)
		}
		for _, viewer := range []string{"", g.Seats[0].ID.String(), SpectatorViewerID} {
			raw, err := json.Marshal(FilterViewFor(v, viewer))
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{`"match"`, `"best_of"`, `"sideboarding"`} {
				if strings.Contains(string(raw), key) {
					t.Errorf("%s: a Commander table's frame contains %s", when, key)
				}
			}
		}
	}
	check("in play")
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	check("ended")
	if v := ViewOfGame(g); v.State != "ended" || v.Outcome == nil {
		t.Errorf("an ended Commander table: state %q, outcome %+v", v.State, v.Outcome)
	}
}

// A best-of-one Standard table has a match of one game, and its end is
// the table's end with the game's own outcome, as before.
func TestBestOfOneViewEndsWithItsGame(t *testing.T) {
	g := standardTable(t, 1)
	v := ViewOfGame(g)
	if v.Match == nil || v.Match.BestOf != 1 || v.Settings.BestOf != 0 {
		t.Fatalf("match %+v, settings.best_of %d", v.Match, v.Settings.BestOf)
	}
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	v = ViewOfGame(g)
	if v.State != "ended" || v.Outcome == nil || v.Outcome.Cause != game.OutcomeCauseLastStanding ||
		v.Outcome.WinnerSeat == nil || *v.Outcome.WinnerSeat != 0 {
		t.Fatalf("state %q, outcome %+v", v.State, v.Outcome)
	}
	if !v.Match.Over || v.Match.Sideboarding != nil || v.Match.WinnerSeat == nil || *v.Match.WinnerSeat != 0 {
		t.Errorf("match = %+v", v.Match)
	}
}
