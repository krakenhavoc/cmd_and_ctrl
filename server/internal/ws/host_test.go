package ws

import (
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
)

func TestPickHost(t *testing.T) {
	a, b, c, d := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	seats := func(mods ...func([]hostSeat)) []hostSeat {
		s := []hostSeat{{id: a}, {id: b, bot: true}, {id: c}, {id: d}}
		for _, m := range mods {
			m(s)
		}
		return s
	}
	out := func(i int) func([]hostSeat) { return func(s []hostSeat) { s[i].eliminated = true } }

	cases := []struct {
		name       string
		designated uuid.UUID
		seats      []hostSeat
		want       uuid.UUID
	}{
		{"nobody designated", uuid.Nil, seats(), uuid.Nil},
		{"designated and present", c, seats(), c},
		{"skips the bot", a, seats(out(0)), c},
		{"skips a departed seat", a, seats(out(0), out(2)), d},
		{"wraps around", d, seats(out(3)), a},
		{"no human left", a, seats(out(0), out(2), out(3)), uuid.Nil},
		{"designated not seated walks from seat 0", uuid.New(), seats(), a},
		{"a bot designated never hosts", b, seats(), c},
		{"empty table", a, nil, uuid.Nil},
	}
	for _, tc := range cases {
		if got := pickHost(tc.designated, tc.seats); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestRoomStampsAndGatesTheHost(t *testing.T) {
	g := seedTestGame(t)
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	p1, p2 := g.Seats[0].ID, g.Seats[1].ID

	view, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range view.Seats {
		if s.IsHost {
			t.Errorf("seat %s is host before anyone was designated", s.Name)
		}
	}

	room.SetHost(p1)
	view, _, _ = room.Snapshot()
	if !view.Seats[0].IsHost || view.Seats[1].IsHost {
		t.Errorf("is_host = %v/%v, want true/false", view.Seats[0].IsHost, view.Seats[1].IsHost)
	}

	cases := []struct {
		name string
		b    Binding
		want bool
	}{
		{"admin", Binding{GameID: g.ID, Admin: true}, true},
		{"admin on a player's seat", Binding{GameID: g.ID, PlayerID: p2, Admin: true}, true},
		{"host", Binding{GameID: g.ID, PlayerID: p1}, true},
		{"other seat", Binding{GameID: g.ID, PlayerID: p2}, false},
		{"spectator", Binding{GameID: g.ID, ReadOnly: true}, false},
		{"read-only on the host seat", Binding{GameID: g.ID, PlayerID: p1, ReadOnly: true}, false},
		{"host id bound to another game", Binding{GameID: uuid.New(), PlayerID: p1}, false},
	}
	for _, tc := range cases {
		if got := room.CanManageTable(tc.b); got != tc.want {
			t.Errorf("%s: CanManageTable = %v, want %v", tc.name, got, tc.want)
		}
	}

	// The host concedes: hosting passes, sticky, and the next capture
	// says so.
	if _, _, err := room.Apply(p1, func() error { return g.Concede(p1) }); err != nil {
		t.Fatal(err)
	}
	if h := room.HostPlayerID(); h != uuid.Nil {
		// Two seats: P1 left, P2 is the only human — but the game has
		// ended with P2 the winner; P2 is still an eligible host.
		if h != p2 {
			t.Errorf("host after concede = %v, want p2", h)
		}
	} else {
		t.Error("no host after the host conceded, want p2")
	}
	if !room.IsHost(p2) || room.IsHost(p1) {
		t.Error("IsHost did not follow the pass")
	}
}
