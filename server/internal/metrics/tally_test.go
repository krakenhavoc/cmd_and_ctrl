package metrics

import (
	"reflect"
	"testing"
)

// Tally is the one definition of the tables collector's numbers (ADR
// 0124 §5): the collector reports it, and the admin views' Live now
// totals are computed with it. Each case is one rule of the counting.
func TestTally(t *testing.T) {
	human := func(p byte, signedIn bool) Seat { return Seat{Player: key(p), Kind: SeatHuman, SignedIn: signedIn} }
	seat := func(g, p byte) Socket { return Socket{Game: key(g), Player: key(p), Role: RoleSeat} }

	cases := []struct {
		name    string
		tables  []Table
		sockets []Socket
		want    Counts
		players int
	}{
		{
			name: "nothing",
			want: Counts{Games: map[GamesKey]int{}, Seats: map[string]int{}, Connected: map[ConnectedKey]int{}, Roles: map[string]int{}},
		},
		{
			name:    "a connected guest and a signed-in person without a socket",
			tables:  []Table{{Game: key(1), State: "active", Seats: []Seat{human(11, false), human(12, true)}}},
			sockets: []Socket{seat(1, 11)},
			want: Counts{
				Games:     map[GamesKey]int{{"active", false}: 1},
				Seats:     map[string]int{SeatHuman: 2},
				Connected: map[ConnectedKey]int{{SeatHuman, AccountGuest}: 1},
				Roles:     map[string]int{RoleSeat: 1},
			},
			players: 1,
		},
		{
			name:    "two tabs on one seat are one connected seat",
			tables:  []Table{{Game: key(1), State: "active", Seats: []Seat{human(11, true)}}},
			sockets: []Socket{seat(1, 11), seat(1, 11)},
			want: Counts{
				Games:     map[GamesKey]int{{"active", false}: 1},
				Seats:     map[string]int{SeatHuman: 1},
				Connected: map[ConnectedKey]int{{SeatHuman, AccountSignedIn}: 1},
				Roles:     map[string]int{RoleSeat: 2},
			},
			players: 1,
		},
		{
			name: "a bot is a seat and never connected; an agent connects like a person",
			tables: []Table{{Game: key(1), State: "active", Seats: []Seat{
				{Player: key(13), Kind: SeatBot}, {Player: key(14), Kind: SeatAgent},
			}}},
			sockets: []Socket{seat(1, 13), seat(1, 14)},
			want: Counts{
				Games:     map[GamesKey]int{{"active", false}: 1},
				Seats:     map[string]int{SeatBot: 1, SeatAgent: 1},
				Connected: map[ConnectedKey]int{{SeatAgent, AccountGuest}: 1},
				Roles:     map[string]int{RoleSeat: 2},
			},
			players: 1,
		},
		{
			name: "archived, waiting and ended tables count as games and nothing else",
			tables: []Table{
				{Game: key(2), State: "active", Archived: true, Seats: []Seat{human(21, false)}},
				{Game: key(3), State: "lobby", Seats: []Seat{human(31, false)}},
				{Game: key(4), State: "ended", Seats: []Seat{human(41, false)}},
			},
			sockets: []Socket{seat(2, 21), seat(3, 31), seat(4, 41)},
			want: Counts{
				Games:     map[GamesKey]int{{"active", true}: 1, {"lobby", false}: 1, {"ended", false}: 1},
				Seats:     map[string]int{},
				Connected: map[ConnectedKey]int{},
				Roles:     map[string]int{RoleSeat: 3},
			},
		},
		{
			name:    "a practice table counts only as a practice table",
			tables:  []Table{{Game: key(5), State: "active", Practice: true, Seats: []Seat{human(51, true)}}},
			sockets: []Socket{seat(5, 51)},
			want: Counts{
				Games:     map[GamesKey]int{},
				Practice:  1,
				Seats:     map[string]int{},
				Connected: map[ConnectedKey]int{},
				Roles:     map[string]int{RoleSeat: 1},
			},
		},
		{
			name:    "spectators and admins count by role, wherever they are",
			tables:  []Table{{Game: key(1), State: "active", Seats: []Seat{human(11, false)}}},
			sockets: []Socket{{Game: key(1), Role: RoleSpectator}, {Game: key(9), Role: RoleSpectator}, {Game: key(1), Role: RoleAdmin}},
			want: Counts{
				Games:     map[GamesKey]int{{"active", false}: 1},
				Seats:     map[string]int{SeatHuman: 1},
				Connected: map[ConnectedKey]int{},
				Roles:     map[string]int{RoleSpectator: 2, RoleAdmin: 1},
			},
		},
		{
			name:    "a seat socket counts only at its own table",
			tables:  []Table{{Game: key(1), State: "active", Seats: []Seat{human(11, false)}}},
			sockets: []Socket{seat(9, 11)},
			want: Counts{
				Games:     map[GamesKey]int{{"active", false}: 1},
				Seats:     map[string]int{SeatHuman: 1},
				Connected: map[ConnectedKey]int{},
				Roles:     map[string]int{RoleSeat: 1},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Tally(tc.tables, tc.sockets)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Tally =\n  %+v\nwant\n  %+v", got, tc.want)
			}
			if p := got.PlayersConnected(); p != tc.players {
				t.Errorf("PlayersConnected = %d, want %d", p, tc.players)
			}
		})
	}
}
