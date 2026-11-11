package accountdelete

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// TestDeleteRemovesThePersonsCubes is ADR 0222 §8 under ADR 0149: a
// deleted person's saved cubes go, and another person's stay.
func TestDeleteRemovesThePersonsCubes(t *testing.T) {
	f := seed(t)
	u, o := f.user.String(), f.other.String()
	exec(t, f.d, `INSERT INTO cubes (id, user_id, name, singleton, list, card_count, created_at, updated_at) VALUES
		(?, ?, 'Vintage', 1, '1 Card A', 1, 5, 5),
		(?, ?, 'Pauper', 0, '2 Card B', 2, 5, 5),
		(?, ?, 'Vintage', 1, '1 Card C', 1, 5, 5)`,
		uuid.NewString(), u, uuid.NewString(), u, uuid.NewString(), o)

	res, err := NewSQLStore(f.d).Delete(context.Background(), f.user)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if res.Counts.Cubes != 2 {
		t.Errorf("counts.Cubes = %d, want 2", res.Counts.Cubes)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM cubes WHERE user_id = ?`, u); n != 0 {
		t.Errorf("the deleted person still has %d cubes", n)
	}
	if n := count(t, f.d, `SELECT COUNT(*) FROM cubes WHERE user_id = ? AND name = 'Vintage' AND list = '1 Card C'`, o); n != 1 {
		t.Errorf("the other person's cube was not kept (%d rows)", n)
	}
}
