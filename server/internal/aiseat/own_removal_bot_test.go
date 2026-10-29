package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #4319: the real catalog and enumerator must carry the declared return
// through to scoring. Inspect every offered cast at our Sol Ring, since
// the enumerator may offer more than one way to pay for the spell.
func TestHeuristicHoldsGenerousGiftRatherThanDestroyItsSolRing(t *testing.T) {
	g := newSettledTable(t, 4319)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	ring := uuid.New()
	g.WithWriteLock(func() {
		for _, s := range g.Seats {
			s.Hand.Cards = nil
			s.Command.Cards = nil
		}
		g.Battlefield.PushTop(game.Card{InstanceID: ring, Name: "Sol Ring",
			OracleID: "6ad8011d-3471-4369-9d68-b264cc027487", TypeLine: "Artifact", ManaCost: "{1}",
			Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		for i := 0; i < 3; i++ {
			g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
				Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		}
		me.Hand.Cards = []game.Card{{InstanceID: uuid.New(), Name: "Generous Gift",
			OracleID: "fae37e28-e137-4177-b973-fa8b4dd8f409", TypeLine: "Instant", ManaCost: "{2}{W}",
			Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}}}
	})
	g.BumpLayerVersionForTest()
	var moves []legal.Move
	casts := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindPass {
			moves = append(moves, m)
		}
		if m.Kind == legal.KindCast {
			ids := targetIDs(t, m)
			if len(ids) == 1 && ids[0] == ring {
				moves = append(moves, m)
				casts++
			}
		}
	}
	if casts == 0 || len(moves) == casts {
		t.Fatalf("need a pass and legal casts at Sol Ring; got %d casts, %d moves", casts, len(moves))
	}
	in := aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: moves}
	for _, enabled := range []bool{true, false} {
		cfg := heuristic.DefaultConfig()
		cfg.NetRemoval = enabled
		d, err := heuristic.NewWithConfig(cfg).Decide(context.Background(), in)
		if err != nil {
			t.Fatal(err)
		}
		want := legal.KindCast
		if enabled {
			want = legal.KindPass
		}
		if moves[d.Index].Kind != want {
			t.Errorf("NetRemoval=%v: chose %s, want %s", enabled, moves[d.Index].Label, want)
		}
	}
}
