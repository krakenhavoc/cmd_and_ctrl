package lobby

import (
	"database/sql"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestGamesOutcomeFollowsTheResult — games.outcome (migration 0007,
// ADR 0057 Decision 7, #1520) tells a win from a draw from a table an
// admin closed, which winner_seat alone cannot: a draw and an
// abandoned table are both NULL there. The game ends over the room,
// the way a WS action does, with nobody calling the lobby.
func TestGamesOutcomeFollowsTheResult(t *testing.T) {
	cases := []struct {
		name string
		end  func(g *game.Game, alice, bob uuid.UUID) error
		want sql.NullString
		seat bool // winner_seat is set
	}{
		{
			name: "win",
			end:  func(g *game.Game, alice, _ uuid.UUID) error { return g.Concede(alice) },
			want: sql.NullString{String: "win", Valid: true},
			seat: true,
		},
		{
			name: "draw",
			end: func(g *game.Game, _, _ uuid.UUID) error {
				g.WithWriteLock(func() {
					g.Outcome = &game.GameOutcome{Kind: game.OutcomeDraw, Cause: game.OutcomeCauseAllLost}
				})
				g.End()
				return nil
			},
			want: sql.NullString{String: "draw", Valid: true},
		},
		{
			name: "closed by an admin",
			end:  func(g *game.Game, _, _ uuid.UUID) error { g.End(); return nil },
			want: sql.NullString{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			l, _ := newDurableLobby(t, dir)
			meta, alice, bob := startTwoSeatGame(t, l, "Outcome")
			d := openTestDB(t, dir)

			read := func() (state string, outcome sql.NullString, winner *int64) {
				t.Helper()
				if err := d.QueryRow(`SELECT state, outcome, winner_seat FROM games WHERE id = ?`,
					meta.ID.String()).Scan(&state, &outcome, &winner); err != nil {
					t.Fatal(err)
				}
				return
			}
			if state, outcome, _ := read(); state != "active" || outcome.Valid {
				t.Fatalf("after Start: state=%q outcome=%v, want active and NULL", state, outcome)
			}

			room := l.RoomOf(meta.ID)
			if _, _, err := room.Apply(alice, func() error { return tc.end(room.Game, alice, bob) }); err != nil {
				t.Fatalf("ending the game: %v", err)
			}
			waitFor(t, func() bool {
				s, _, _ := read()
				return s == "ended"
			})
			_, outcome, winner := read()
			if outcome != tc.want {
				t.Errorf("outcome = %v, want %v", outcome, tc.want)
			}
			if (winner != nil) != tc.seat {
				t.Errorf("winner_seat = %v, want set=%v", winner, tc.seat)
			}

			// /me/games and a reload read it back from the row.
			ctx, cancel := storeCtx()
			defer cancel()
			rec, _, err := l.store.LoadGame(ctx, meta.ID)
			if err != nil {
				t.Fatal(err)
			}
			if rec.Outcome != tc.want.String {
				t.Errorf("LoadGame outcome = %q, want %q", rec.Outcome, tc.want.String)
			}
		})
	}
}
