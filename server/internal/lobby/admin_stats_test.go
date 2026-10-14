package lobby

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
)

// GET /admin/stats serves the overview to an admin. The gate it sits
// behind is checked with the other admin routes in admins_test.go.
func TestAdminStatsServesTheOverview(t *testing.T) {
	w := newAdminViewsWorld(t)
	w.l.SetGameStats(gamestats.NewSQLStore(w.d))
	// The seeded "Ended table" has no record, so the admin is told to
	// run the backfill.
	code, raw := w.get(t, "/admin/stats")
	if code != http.StatusOK {
		t.Fatalf("GET /admin/stats: %d %s", code, raw)
	}
	var ov gamestats.Overview
	if err := json.Unmarshal(raw, &ov); err != nil {
		t.Fatal(err)
	}
	if ov.Totals.Games != 0 || ov.Totals.Unrecorded != 1 || ov.Recent == nil || ov.People == nil || ov.Cards == nil {
		t.Errorf("overview = %s", raw)
	}
	var m map[string]json.RawMessage
	_ = json.Unmarshal(raw, &m)
	for _, f := range []string{"totals", "recent", "people", "commanders_by_games", "commanders_by_wins", "cards"} {
		if _, ok := m[f]; !ok {
			t.Errorf("response has no %q", f)
		}
	}
}

// With no stats store wired the route still answers, empty.
func TestAdminStatsWithoutAStore(t *testing.T) {
	w := newAdminViewsWorld(t)
	if code, raw := w.get(t, "/admin/stats"); code != http.StatusOK {
		t.Fatalf("GET /admin/stats: %d %s", code, raw)
	}
}
