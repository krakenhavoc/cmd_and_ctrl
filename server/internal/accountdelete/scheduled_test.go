package accountdelete

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// TestDeleteRemovesScheduledTablesAndRSVPs is ADR 0215's privacy
// section: the plans the person hosts go with their lists (an upcoming
// one is cancelled by it), every RSVP of theirs goes, by account and by
// Discord id, a plan still waiting deals its seats again, and other
// people's rows stay.
func TestDeleteRemovesScheduledTablesAndRSVPs(t *testing.T) {
	f := seed(t)
	u, o := f.user.String(), f.other.String()
	upcoming, opened, theirs, byDiscord := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	plan := `INSERT INTO scheduled_tables (id, host_user_id, name, starts_at, length_minutes, zone, players, status, game_id, created_at, updated_at)
		VALUES (?, ?, 'P', 10, 180, 'UTC', ?, ?, ?, 1, 1)`
	exec(t, f.d, plan, upcoming, u, 4, "scheduled", nil)
	exec(t, f.d, plan, opened, u, 4, "open", f.otherGame.String())
	exec(t, f.d, plan, theirs, o, 2, "scheduled", nil)
	exec(t, f.d, plan, byDiscord, o, 4, "scheduled", nil)
	rsvp := `INSERT INTO scheduled_rsvps (schedule_id, discord_id, user_id, source, state, rsvp_at, held_until, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1)`
	exec(t, f.d, rsvp, upcoming, userSubject, u, "host", "seat", 1, nil)
	exec(t, f.d, rsvp, upcoming, otherSubject, o, "site", "seat", 2, nil)
	exec(t, f.d, rsvp, opened, userSubject, u, "host", "held", 1, 99)
	// On the other person's two-seat plan the deleted person holds the
	// second seat and a third person waits.
	exec(t, f.d, rsvp, theirs, otherSubject, o, "host", "seat", 1, nil)
	exec(t, f.d, rsvp, theirs, userSubject, u, "site", "seat", 2, nil)
	exec(t, f.d, rsvp, theirs, "333333333333333333", nil, "discord", "waiting", 3, nil)
	// A row by the person's Discord id with no account on it.
	exec(t, f.d, rsvp, byDiscord, otherSubject, o, "host", "seat", 1, nil)
	exec(t, f.d, rsvp, byDiscord, userSubject, nil, "discord", "seat", 2, nil)

	res, err := NewSQLStore(f.d).Delete(context.Background(), f.user)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	c := res.Counts
	if c.ScheduledPlans != 2 || c.ScheduledCancelled != 1 || c.ScheduledRSVPs != 2 {
		t.Errorf("counts: plans %d, cancelled %d, rsvps %d; want 2, 1, 2", c.ScheduledPlans, c.ScheduledCancelled, c.ScheduledRSVPs)
	}
	if len(res.ScheduledGames) != 1 || res.ScheduledGames[0] != f.otherGame.String() {
		t.Errorf("ScheduledGames = %v, want the opened plan's table", res.ScheduledGames)
	}
	for q, want := range map[string]int{
		`SELECT COUNT(*) FROM scheduled_tables WHERE host_user_id = ?`:                                 0,
		`SELECT COUNT(*) FROM scheduled_rsvps WHERE user_id = ? OR discord_id = '` + userSubject + `'`: 0,
	} {
		if n := count(t, f.d, q, u); n != want {
			t.Errorf("%s = %d, want %d", q, n, want)
		}
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM scheduled_tables WHERE host_user_id = ?`, o); n != 2 {
		t.Errorf("the other person's plans: %d, want 2", n)
	}
	var state string
	if err := f.d.QueryRow(`SELECT state FROM scheduled_rsvps WHERE schedule_id = ? AND discord_id = '333333333333333333'`, theirs).Scan(&state); err != nil || state != "seat" {
		t.Errorf("the waiting person on the other plan: %q %v, want moved up to a seat", state, err)
	}
}
