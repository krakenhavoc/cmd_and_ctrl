package accountdelete

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// TestDeletePassesOnOwnedPodsAndRemovesThePersonsPodRows is ADR 0220 §12:
// a pod the person owns passes to the member who joined first, or is
// deleted if they were the last; their member rows, every invite naming
// them as invitee or inviter, and their dismissals go; the pod's name
// stays with the pod, and other people's rows stay.
func TestDeletePassesOnOwnedPodsAndRemovesThePersonsPodRows(t *testing.T) {
	f := seed(t)
	third := uuid.New()
	exec(t, f.d, `INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'Cy', 1, 1)`, third.String())
	u, o, c := f.user.String(), f.other.String(), third.String()
	shared, alone, theirs := uuid.NewString(), uuid.NewString(), uuid.NewString()
	exec(t, f.d, `INSERT INTO pods (id, name, owner_id, invite_hash, invite_expires_at, created_at, updated_at)
		VALUES (?, 'Shared', ?, 'h', 99, 1, 1), (?, 'Alone', ?, NULL, NULL, 1, 1), (?, 'Theirs', ?, NULL, NULL, 1, 1)`,
		shared, u, alone, u, theirs, o)
	// In Shared, Cy joined before the other person, so Cy takes it.
	exec(t, f.d, `INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES
		(?, ?, 1), (?, ?, 3), (?, ?, 2), (?, ?, 1), (?, ?, 1), (?, ?, 2)`,
		shared, u, shared, o, shared, c, alone, u, theirs, o, theirs, u)
	exec(t, f.d, `INSERT INTO pod_invites (pod_id, user_id, invited_by, created_at) VALUES
		(?, ?, ?, 1), (?, ?, ?, 1), (?, ?, ?, 1)`,
		theirs, c, u, // sent by the person
		alone, u, o, // sent to the person
		shared, o, c) // between two others
	exec(t, f.d, `INSERT INTO pod_suggestion_dismissals (user_id, group_key, dismissed_at) VALUES (?, 'k', 1), (?, 'k', 1)`, u, o)

	res, err := NewSQLStore(f.d).Delete(context.Background(), f.user)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got := res.Counts
	if got.PodsPassedOn != 1 || got.PodsDeleted != 1 || got.PodMembers != 3 || got.PodInvites != 2 || got.PodDismissals != 1 {
		t.Errorf("counts = passed %d, deleted %d, members %d, invites %d, dismissals %d; want 1, 1, 3, 2, 1",
			got.PodsPassedOn, got.PodsDeleted, got.PodMembers, got.PodInvites, got.PodDismissals)
	}
	var owner, name string
	if err := f.d.QueryRow(`SELECT owner_id, name FROM pods WHERE id = ?`, shared).Scan(&owner, &name); err != nil {
		t.Fatal(err)
	}
	if owner != c || name != "Shared" {
		t.Errorf("Shared: owner %s name %q; want Cy, who joined first, and the name kept", owner, name)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM pods WHERE id = ?`, alone); n != 0 {
		t.Error("a pod whose last member was deleted is still there")
	}
	for q, want := range map[string]int{
		`SELECT COUNT(*) FROM pod_members WHERE user_id = ? OR user_id = ?`:               0,
		`SELECT COUNT(*) FROM pod_invites WHERE user_id = ? OR invited_by = ?`:            0,
		`SELECT COUNT(*) FROM pod_suggestion_dismissals WHERE user_id = ? OR user_id = ?`: 0,
		`SELECT COUNT(*) FROM pods WHERE owner_id = ? OR owner_id = ?`:                    0,
	} {
		if n := count(t, f.d, q, u, u); n != want {
			t.Errorf("%s = %d, want %d", q, n, want)
		}
	}
	// The owner is always a member (ADR 0220 §9).
	if n := count(t, f.d, `SELECT COUNT(*) FROM pods p WHERE NOT EXISTS
		(SELECT 1 FROM pod_members m WHERE m.pod_id = p.id AND m.user_id = p.owner_id)`); n != 0 {
		t.Errorf("%d pods have an owner who is not a member", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM pod_invites WHERE pod_id = ? AND user_id = ?`, shared, o); n != 1 {
		t.Errorf("an invite between two others went: %d", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM pod_suggestion_dismissals WHERE user_id = ?`, o); n != 1 {
		t.Errorf("another person's dismissal went: %d", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM pod_members WHERE pod_id = ?`, theirs); n != 1 {
		t.Errorf("Theirs keeps %d members, want its owner", n)
	}
}
