package accountdelete

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// TestDeleteRemovesThePersonsDraftHistory is ADR 0222 §9 under ADR
// 0149: a deleted person's finished drafts and their rounds go, a round
// of another person's draft that named them as the opponent names
// nobody, and the other person's own record stays.
func TestDeleteRemovesThePersonsDraftHistory(t *testing.T) {
	f := seed(t)
	u, o := f.user.String(), f.other.String()
	mine, theirs := uuid.NewString(), uuid.NewString()
	exec(t, f.d, `INSERT INTO draft_history (draft_id, user_id, draft_name, set_code, set_name, mode, seats, best_of, seat, drafted_at, picks, basics) VALUES
		(?, ?, 'A draft', 'blb', 'Bloomburrow', 'pick_one', 8, 1, 0, 5, '[{"name":"Card A"}]', '{"Forest":17}'),
		(?, ?, 'A draft', 'blb', 'Bloomburrow', 'pick_one', 8, 1, 1, 5, '[{"name":"Card B"}]', '{}')`,
		mine, u, theirs, o)
	exec(t, f.d, `INSERT INTO draft_history_matches (draft_id, user_id, round, opponent_kind, opponent_user_id, result, games_won, games_lost) VALUES
		(?, ?, 1, 'person', ?, 'win', 1, 0),
		(?, ?, 2, 'bot', NULL, 'loss', 0, 1),
		(?, ?, 1, 'person', ?, 'loss', 0, 1)`,
		mine, u, o, mine, u, theirs, o, u)

	res, err := NewSQLStore(f.d).Delete(context.Background(), f.user)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if c := res.Counts; c.Drafts != 1 || c.DraftMatches != 2 || c.DraftOpponents != 1 {
		t.Errorf("counts = drafts %d, rounds %d, opponents %d; want 1, 2, 1", c.Drafts, c.DraftMatches, c.DraftOpponents)
	}
	for q, want := range map[string]int{
		`SELECT COUNT(*) FROM draft_history WHERE user_id = ?`:                  0,
		`SELECT COUNT(*) FROM draft_history_matches WHERE user_id = ?`:          0,
		`SELECT COUNT(*) FROM draft_history_matches WHERE opponent_user_id = ?`: 0,
	} {
		if n := count(t, f.d, q, u); n != want {
			t.Errorf("%s = %d, want %d", q, n, want)
		}
	}
	// The other person's draft and its round are still theirs.
	if n := count(t, f.d, `SELECT COUNT(*) FROM draft_history WHERE user_id = ?`, o); n != 1 {
		t.Errorf("the other person has %d draft records, want 1", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM draft_history_matches WHERE user_id = ? AND opponent_kind = 'person' AND opponent_user_id IS NULL AND result = 'loss'`, o); n != 1 {
		t.Errorf("the other person's round was not kept with its opponent cleared (%d rows)", n)
	}
}
