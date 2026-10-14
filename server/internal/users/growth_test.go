package users

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// The users store's half of the growth collector (#3557): new vs
// returning, week-over-week retention, signup cohorts and the finished
// games' engagement sums, over real rows.
func TestGrowth(t *testing.T) {
	s, d := openStore(t, nil)
	now := time.Now().UTC()

	user := func(createdAgo time.Duration) string {
		id := uuid.NewString()
		at := now.Add(-createdAgo).UnixMilli()
		if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'p', ?, ?)`, id, at, at); err != nil {
			t.Fatal(err)
		}
		return id
	}
	// game seats the users at a table that ended endedAgo ago, with
	// statistics when turns > 0.
	game := func(endedAgo time.Duration, turns, rounds int, users ...string) {
		id := uuid.NewString()
		end := now.Add(-endedAgo).UnixMilli()
		if _, err := d.Exec(`INSERT INTO games (id, name, state, created_at, started_at, ended_at) VALUES (?, 'g', 'ended', 1, ?, ?)`,
			id, end-int64(30*60*1000), end); err != nil {
			t.Fatal(err)
		}
		for i, u := range users {
			var user any
			if u != "" {
				user = u
			}
			if _, err := d.Exec(`INSERT INTO seats (game_id, seat, player_id, user_id) VALUES (?, ?, ?, ?)`, id, i, uuid.NewString(), user); err != nil {
				t.Fatal(err)
			}
		}
		if turns > 0 {
			if _, err := d.Exec(`INSERT INTO game_stats (game_id, stats_version, source, log_complete, turns, rounds, computed_at) VALUES (?, 1, 'live', 1, ?, ?, 1)`,
				id, turns, rounds); err != nil {
				t.Fatal(err)
			}
		}
	}

	fresh := user(2 * day)     // created 2 days ago: in no cohort
	veteran := user(100 * day) // played 20 days ago and 1 day ago: returning
	lapsed := user(100 * day)  // played 10 days ago only: in the previous week, not retained
	kept := user(100 * day)    // played 10 days ago and 2 days ago: retained
	cohort1 := user(10 * day)  // created 10 days ago (1w band), played 3 days ago
	user(11 * day)             // 1w band, never played
	cohort4 := user(31 * day)  // 4w band, played 1 day ago
	user(60 * day)             // 8w band, idle

	game(20*day, 40, 10, veteran)
	game(day-time.Hour, 30, 8, veteran, "", cohort4)
	game(10*day, 20, 5, lapsed, kept)
	game(2*day, 25, 6, kept, fresh)
	game(3*day, 0, 0, cohort1) // no statistics: not an engaged game

	got, err := s.Growth(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}

	// 7d: veteran, kept, fresh, cohort1 and cohort4 played. The last
	// three first played inside the window; veteran and kept did before.
	w7 := got.Windows["7d"]
	if w7.NewPlayers != 3 || w7.ReturningPlayers != 2 {
		t.Errorf("7d new/returning = %d/%d, want 3/2", w7.NewPlayers, w7.ReturningPlayers)
	}
	w1 := got.Windows["1d"]
	if w1.NewPlayers != 1 || w1.ReturningPlayers != 1 {
		t.Errorf("1d new/returning = %d/%d, want 1/1", w1.NewPlayers, w1.ReturningPlayers)
	}
	// Everyone who played in 30 days first played inside it.
	w30 := got.Windows["30d"]
	if w30.NewPlayers != 6 || w30.ReturningPlayers != 0 {
		t.Errorf("30d new/returning = %d/%d, want 6/0", w30.NewPlayers, w30.ReturningPlayers)
	}

	// Retention: played 14..7 days ago: lapsed, kept. Of them, kept
	// played in the last 7.
	if got.RetentionPrevious != 2 || got.RetentionRetained != 1 {
		t.Errorf("retention = %d of %d, want 1 of 2", got.RetentionRetained, got.RetentionPrevious)
	}

	// Cohorts.
	if c := got.Cohorts["1w"]; c.Accounts != 2 || c.Active != 1 {
		t.Errorf("1w cohort = %+v, want 2 accounts, 1 active", c)
	}
	if c := got.Cohorts["4w"]; c.Accounts != 1 || c.Active != 1 {
		t.Errorf("4w cohort = %+v, want 1 account, 1 active", c)
	}
	if c := got.Cohorts["8w"]; c.Accounts != 1 || c.Active != 0 {
		t.Errorf("8w cohort = %+v, want 1 account, 0 active", c)
	}

	// Engagement in 7d: the 1d-ago and 2d-ago games have statistics (the
	// 3d one has none): 2 games, 3+2 seats, 55 turns, 14 rounds, and the
	// signed-in accounts veteran, cohort4, kept, fresh.
	if w7.GamesFinished != 2 || w7.Seats != 5 || w7.Turns != 55 || w7.Rounds != 14 || w7.FinishedPlayers != 4 {
		t.Errorf("7d engagement = %+v, want 2 games, 5 seats, 55 turns, 14 rounds, 4 players", w7)
	}
}
