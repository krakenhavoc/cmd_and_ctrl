package accountdelete

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// TestDeleteClearsFriendsAndBlocksOnBothSides is ADR 0213 §10: every
// friendship, friend request and personal block that names the person
// goes, whichever side they are on, and so does their settings row; the
// same rows between two other people stay.
func TestDeleteClearsFriendsAndBlocksOnBothSides(t *testing.T) {
	f := seed(t)
	third := uuid.New()
	exec(t, f.d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'Cy', 1, 1)`, third.String())
	u, o, c := f.user.String(), f.other.String(), third.String()
	ordered := func(a, b string) (string, string) {
		if a < b {
			return a, b
		}
		return b, a
	}
	for _, p := range [][2]string{{u, o}, {u, c}, {o, c}} {
		a, b := ordered(p[0], p[1])
		exec(t, f.d, `INSERT INTO friendships (user_a, user_b, created_at) VALUES (?, ?, 1)`, a, b)
	}
	// Sent and received, ignored or not, and one between the others.
	exec(t, f.d, `INSERT INTO friend_requests (from_user, to_user, created_at, ignored_at)
		VALUES (?, ?, 1, NULL), (?, ?, 1, 2), (?, ?, 1, NULL)`, u, o, c, u, o, c)
	exec(t, f.d, `INSERT INTO user_blocks (user_id, blocked_user_id, created_at) VALUES (?, ?, 1), (?, ?, 1), (?, ?, 1)`,
		u, c, o, u, c, o)
	exec(t, f.d, `INSERT INTO social_prefs (user_id, presence, requests, updated_at)
		VALUES (?, 'nobody', 'nobody', 1), (?, 'played_with', 'played_with', 1)`, u, o)

	res, err := NewSQLStore(f.d).Delete(context.Background(), f.user)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	for q, want := range map[string]int{
		`SELECT COUNT(*) FROM friendships WHERE user_a = ? OR user_b = ?`:              0,
		`SELECT COUNT(*) FROM friend_requests WHERE from_user = ? OR to_user = ?`:      0,
		`SELECT COUNT(*) FROM user_blocks WHERE user_id = ? OR blocked_user_id = ?`:    0,
		`SELECT COUNT(*) FROM social_prefs WHERE user_id = ? OR user_id = ?`:           0,
		`SELECT COUNT(*) FROM friendships WHERE user_a <> ? AND user_b <> ?`:           1,
		`SELECT COUNT(*) FROM friend_requests WHERE from_user <> ? AND to_user <> ?`:   1,
		`SELECT COUNT(*) FROM user_blocks WHERE user_id <> ? AND blocked_user_id <> ?`: 1,
	} {
		if n := count(t, f.d, q, u, u); n != want {
			t.Errorf("%s = %d, want %d", q, n, want)
		}
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM social_prefs WHERE user_id = ?`, o); n != 1 {
		t.Errorf("the other person's settings went: %d", n)
	}
	if got := res.Counts; got.Friendships != 2 || got.FriendRequests != 2 || got.UserBlocks != 2 || got.SocialPrefRows != 1 {
		t.Errorf("counts = friendships %d, requests %d, blocks %d, settings %d; want 2, 2, 2, 1",
			got.Friendships, got.FriendRequests, got.UserBlocks, got.SocialPrefRows)
	}
}
