package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// plot_special_action_view_test.go — #1342: a card with plot in the
// viewer's own hand carries a `plot` special-action row, and its
// `available` flag is the engine's CR 702.170a window — the owner's
// main phase with the stack empty. The row is the client's plot
// button; no new field, one new kind on the existing surface.
func TestPlotRowIsProjectedWithTheMainPhaseWindow(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	const oracle = "protocol-test-plot-card"
	installSpecialAction(t, oracle, game.SpecialAction{
		Kind: game.SpecialActionPlot, Cost: "{3}{U}", Label: "Plot {3}{U}",
	})
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Djinn of Fool's Fall",
		TypeLine:   "Creature — Djinn",
		ManaCost:   "{4}{U}",
		OracleID:   oracle,
		Owner:      me.ID,
		Controller: me.ID,
		KnownBy:    map[uuid.UUID]bool{me.ID: true},
	})

	rowAt := func(step game.Step) SpecialActionView {
		t.Helper()
		g.WithWriteLock(func() { g.Turn.Step = step })
		rows := handCardView(t, FilterViewFor(ViewOfGame(g), me.ID.String()), g.Turn.ActiveSeat, id).SpecialActions
		if len(rows) != 1 {
			t.Fatalf("want exactly one special-action row, got %+v", rows)
		}
		return rows[0]
	}

	row := rowAt(game.StepPrecombatMain)
	if row.Kind != "plot" || row.Label != "Plot {3}{U}" || row.Cost != "{3}{U}" {
		t.Errorf("row = %+v, want kind plot, label and cost {3}{U}", row)
	}
	if !row.Available {
		t.Error("plot row greyed in its owner's main phase with the stack empty")
	}
	if row = rowAt(game.StepUpkeep); row.Available {
		t.Error("plot row available in the upkeep — CR 702.170a is main phase only")
	}
}
