package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_shields_view_test.go — ADR 0106 §4 decision 6 (#1806): a
// seat's live "can't be countered" grants and unspent promises reach
// every viewer as `counter_shields`, and go when they end.
func TestSeatCarriesItsCounterShields(t *testing.T) {
	g := newTwoSeatGame(t)
	a, b := g.Seats[0], g.Seats[1]
	insist, veil := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.GrantCounterShieldForEffect(a.ID, game.CounterShieldGrant{
			Text: "Spells you control can't be countered this turn.",
		}, "Veil of Summer", veil)
		g.GrantCounterShieldForEffect(a.ID, game.CounterShieldGrant{
			NextOnly: true, Filter: game.PermissionFilter{CreatureOnly: true},
			Text: "The next creature spell you cast this turn can't be countered.",
		}, "Insist", insist)
	})
	v := ViewOfGame(g)
	got := v.Seats[0].CounterShields
	if len(got) != 2 {
		t.Fatalf("seat A counter_shields %+v", got)
	}
	if got[0].SourceName != "Veil of Summer" || got[0].NextOnly || got[0].Source != veil.String() ||
		got[0].Text != "Spells you control can't be countered this turn." {
		t.Errorf("the turn grant %+v", got[0])
	}
	if got[1].SourceName != "Insist" || !got[1].NextOnly {
		t.Errorf("the promise %+v", got[1])
	}
	if len(v.Seats[1].CounterShields) != 0 {
		t.Errorf("seat B has none: %+v", v.Seats[1].CounterShields)
	}
	// Public: an opponent's filtered view carries it unchanged.
	filtered := FilterViewFor(v, b.ID.String())
	if len(filtered.Seats[0].CounterShields) != 2 {
		t.Errorf("an opponent sees the grants: %+v", filtered.Seats[0].CounterShields)
	}
	raw, err := json.Marshal(v.Seats[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"counter_shields":[{"source":"`) || !strings.Contains(string(raw), `"next_only":true`) {
		t.Errorf("wire shape: %s", raw)
	}
	raw, _ = json.Marshal(v.Seats[1])
	if strings.Contains(string(raw), "counter_shields") {
		t.Errorf("omitted for a seat with none: %s", raw)
	}
}
