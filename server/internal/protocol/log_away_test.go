package protocol

// log_away_test.go — ADR 0177: the table moving on without a player who
// is gone is said out loud, to everyone.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func awayLines(g *game.Game, viewer string) []string {
	var out []string
	for _, e := range ViewOfGameFor(g, viewer).Log {
		if e.Kind == LogAwaySeat {
			out = append(out, e.Text)
		}
	}
	return out
}

func TestAwaySeatLinesNameWhoDecidedAndForWhom(t *testing.T) {
	g := newTwoSeatGame(t)
	ana, bo := g.Seats[0], g.Seats[1]
	if err := g.NoteAwaySeat(ana.ID, bo.ID, game.AwayStandIn, 3); err != nil {
		t.Fatal(err)
	}
	if err := g.NoteAwaySeat(bo.ID, bo.ID, game.AwayTakeBack, 0); err != nil {
		t.Fatal(err)
	}
	if err := g.NoteAwaySeat(uuid.Nil, bo.ID, game.AwayStandIn, 12); err != nil {
		t.Fatal(err)
	}
	if err := g.ConcedeForAway(ana.ID, bo.ID, 5); err != nil {
		t.Fatal(err)
	}
	want := []string{
		ana.Name + " handed " + bo.Name + "'s seat to a bot (offline 3m)",
		bo.Name + " took their seat back",
		"The admin handed " + bo.Name + "'s seat to a bot (offline 12m)",
		ana.Name + " conceded for " + bo.Name + " (offline 5m)",
	}
	for _, viewer := range []string{ana.ID.String(), bo.ID.String(), "", SpectatorViewerID} {
		got := awayLines(g, viewer)
		if len(got) != len(want) {
			t.Fatalf("viewer %q: %d away lines %q, want %q", viewer, len(got), got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("viewer %q line %d = %q, want %q", viewer, i, got[i], want[i])
			}
		}
	}
}

func TestAwayAfterSettingIsNarrated(t *testing.T) {
	g := newTwoSeatGame(t)
	host := g.Seats[0]
	for _, v := range []int{1, 0, 7} {
		v := v
		if err := g.UpdateSettings(host.ID, game.SettingsPatch{AwayAfter: &v}); err != nil {
			t.Fatal(err)
		}
	}
	got := texts(settingsLines(t, g, host.ID.String()))
	want := []string{
		host.Name + " (host) set the away time to 1 minute",
		host.Name + " (host) turned off moving on without an away player",
		host.Name + " (host) set the away time to 7 minutes",
	}
	if len(got) != len(want) {
		t.Fatalf("lines %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}
