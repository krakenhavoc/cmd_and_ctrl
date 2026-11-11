package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// discard_land_cost_test.go — #5457: "As an additional cost to cast this
// spell, discard a land card" (Magmatic Insight). The enumerator pays the
// discard only with land cards, and offers no cast without one (#544).

const oracleMagmaticInsight = "6a9031b6-3f4d-4a36-90e1-ac77b8518cee"

func TestDiscardLandCostIsPaidOnlyWithLandCards(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Mountain")
	handCard(seat, game.Card{Name: "Pitch", TypeLine: "Instant", ManaCost: "{1}"})
	insight := handCard(seat, game.Card{Name: "Magmatic Insight", TypeLine: "Sorcery", ManaCost: "{R}", OracleID: oracleMagmaticInsight})

	for _, m := range legal.EnumerateFor(g, seat.ID) {
		if m.Source == insight && m.Kind == legal.KindCast {
			t.Fatalf("a cast was offered with no land card to discard: %s", m.Params)
		}
	}

	land := handCard(seat, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	casts := 0
	for _, m := range legal.EnumerateFor(g, seat.ID) {
		if m.Source != insight || m.Kind != legal.KindCast {
			continue
		}
		casts++
		var p struct {
			DiscardIDs []string `json:"discard_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		if len(p.DiscardIDs) != 1 || p.DiscardIDs[0] != land.String() {
			t.Errorf("discard_ids = %v, want only the land card", p.DiscardIDs)
		}
	}
	if casts == 0 {
		t.Fatalf("no cast offered with a land card in hand")
	}
}
