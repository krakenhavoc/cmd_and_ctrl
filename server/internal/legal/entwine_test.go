package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// entwine_test.go — CR 702.42a, #3866: the enumerator half. An entwine
// spell is its printed count unentwined and every mode entwined, and the
// two are different casts: the entwined one names the entwine cost on
// optional_costs and is offered only when the seat can pay it. Every
// offered move is dispatched against a clone, so "offered" and
// "accepted" cannot drift apart (#544).

const oracleSpectralShift = "99eb4f63-841f-4b6b-9794-82bc2ed7b758"

type entwineOffers struct {
	single, entwined, bothUnentwined, oneEntwined int
}

func countEntwineOffers(t *testing.T, moves []legal.Move) entwineOffers {
	t.Helper()
	var out entwineOffers
	for _, m := range moves {
		var p struct {
			Modes         []int `json:"modes"`
			OptionalCosts []int `json:"optional_costs"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("bad params %s: %v", m.Params, err)
		}
		switch entwined := len(p.OptionalCosts) > 0; {
		case len(p.Modes) == 1 && !entwined:
			out.single++
		case len(p.Modes) == 2 && entwined:
			out.entwined++
		case len(p.Modes) == 2:
			out.bothUnentwined++
		default:
			out.oneEntwined++
		}
	}
	return out
}

func TestEntwineIsOfferedOnlyWithEveryModeAndOnlyWhenPayable(t *testing.T) {
	for _, tc := range []struct {
		islands      int
		wantEntwined bool
	}{
		{2, false}, // {1}{U}: the printed cost, not the {2} more
		{4, true},  // {3}{U}: both
	} {
		g := newTable(t)
		active := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
		clearHand(active)
		spell := handCard(active, game.Card{
			Name: "Spectral Shift", TypeLine: "Instant",
			OracleID: oracleSpectralShift, ManaCost: "{1}{U}",
		})
		battlefieldCard(g, opp, creature("Their Bear", "{1}{G}", 2, 2))
		lands(g, active, "Island", "Island", tc.islands)
		advanceTo(t, g, game.StepPrecombatMain)

		moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell)
		got := countEntwineOffers(t, moves)
		if got.single == 0 {
			t.Errorf("%d Islands: no single-mode cast offered", tc.islands)
		}
		if got.bothUnentwined != 0 || got.oneEntwined != 0 {
			t.Errorf("%d Islands: offered %d unentwined two-mode and %d entwined one-mode casts, want none", tc.islands, got.bothUnentwined, got.oneEntwined)
		}
		if (got.entwined > 0) != tc.wantEntwined {
			t.Errorf("%d Islands: %d entwined casts offered, want any = %v", tc.islands, got.entwined, tc.wantEntwined)
		}
		dispatchAll(t, g, active.ID, moves)
	}
}
