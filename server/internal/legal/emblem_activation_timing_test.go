package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// emblem_activation_timing_test.go — #1275: the per-player activation
// timing statement from an EMBLEM, asked of the two readers that are
// not the engine. The enumerator and the view's `timing_closed` both
// read ActivationTimingOpenLocked, so neither learns the emblem walk;
// this is the test that they did not have to.

const oracleTemporalArchmage = "07d0b06b-80cb-4518-92c9-84ea87a7e08a"

// loyaltyRowsClosed reports, per ability row of `walker` on the wire
// `viewer` receives, whether the engine said the window is shut.
func loyaltyRowsClosed(t *testing.T, g *game.Game, viewer, walker uuid.UUID) []bool {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != walker.String() {
			continue
		}
		out := make([]bool, 0, len(c.ActivatedAbilities))
		for _, row := range c.ActivatedAbilities {
			out = append(out, row.TimingClosed)
		}
		return out
	}
	t.Fatalf("walker %s missing from the battlefield view", walker)
	return nil
}

// Teferi, Temporal Archmage's emblem on an OPPONENT's end step: without
// it a loyalty ability is neither offered nor open on the wire
// (CR 606.3), and with it both readers open the same rows — and what
// the enumerator offers, the engine accepts.
func TestArchmageEmblemOpensLoyaltyOnTheListAndTheWire(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if active.ID == me.ID {
		t.Fatal("setup: need a non-active seat")
	}
	walker := battlefieldCard(g, me, game.Card{
		Name: "Teferi, Time Raveler", TypeLine: "Legendary Planeswalker — Teferi",
		OracleID: oracleTimeRavelerTiming, Counters: map[string]int{"loyalty": 4},
	})
	advanceTo(t, g, game.StepEnd)
	// The active seat passes: I hold priority on somebody else's turn,
	// which is the half of CR 606.3 the emblem's "on any player's
	// turn" is about.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if len(legal.EnumerateFor(g, me.ID)) == 0 {
		t.Fatal("setup: I do not hold priority")
	}

	if acts := activationsOf(legal.EnumerateFor(g, me.ID), walker); len(acts) != 0 {
		t.Fatalf("a loyalty ability offered on an opponent's turn with no emblem: %v", labels(acts))
	}
	rows := loyaltyRowsClosed(t, g, me.ID, walker)
	if len(rows) == 0 {
		t.Fatal("setup: the walker has no ability rows on the wire")
	}
	for i, closed := range rows {
		if !closed {
			t.Errorf("row %d: timing_closed absent with no emblem", i)
		}
	}

	// The −10 resolved earlier; the Archmage is in the graveyard, which
	// is where CreateEmblemForEffect finds its source most of the time.
	archmage := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: archmage, Name: "Teferi, Temporal Archmage",
		TypeLine: "Legendary Planeswalker — Teferi", OracleID: oracleTemporalArchmage,
		Owner: me.ID, Controller: me.ID,
	})
	var err error
	g.WithWriteLock(func() { err = g.CreateEmblemForEffect(me.ID, archmage) })
	if err != nil {
		t.Fatalf("CreateEmblemForEffect: %v", err)
	}

	acts := activationsOf(legal.EnumerateFor(g, me.ID), walker)
	if len(acts) == 0 {
		t.Fatal("the emblem did not put the loyalty abilities on the list")
	}
	dispatchAll(t, g, me.ID, acts)
	for i, closed := range loyaltyRowsClosed(t, g, me.ID, walker) {
		if closed {
			t.Errorf("row %d: timing_closed still set under the emblem", i)
		}
	}
}
