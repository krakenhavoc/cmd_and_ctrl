package protocol

import (
	"testing"
)

// speed_view_test.go — the wire half of ADR 0136 (#2122): a player's
// speed is public on PlayerView.speed, omitted while they have none,
// and every change is a narrated `speed` log line.

func TestPlayerViewCarriesSpeedForEveryViewer(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	g.SetSpeedForTest(me.ID, 3)

	for _, viewer := range []string{me.ID.String(), them.ID.String()} {
		v := ViewOfGameFor(g, viewer)
		var got, other int = -1, -1
		for _, s := range v.Seats {
			if s.ID == me.ID.String() {
				got = s.Speed
			}
			if s.ID == them.ID.String() {
				other = s.Speed
			}
		}
		if got != 3 {
			t.Errorf("viewer %s sees speed %d, want 3", viewer, got)
		}
		if other != 0 {
			t.Errorf("viewer %s sees speed %d for a player with none, want 0 (omitted)", viewer, other)
		}
	}
}

func TestSpeedLogLines(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	g.SetSpeedForTest(me.ID, 2)
	entry := findLog(t, ViewOfGame(g).Log, LogSpeed)
	if entry.Amount != 2 || entry.Seat != 0 {
		t.Errorf("entry amount=%d seat=%d, want 2 and seat 0", entry.Amount, entry.Seat)
	}
	if want := me.Name + "'s speed is now 2"; entry.Text != want {
		t.Errorf("text = %q, want %q", entry.Text, want)
	}

	g.SetSpeedForTest(me.ID, 4)
	var last LogEvent
	for _, e := range ViewOfGame(g).Log {
		if e.Kind == LogSpeed {
			last = e
		}
	}
	if want := me.Name + " has max speed"; last.Text != want {
		t.Errorf("text = %q, want %q", last.Text, want)
	}
}
