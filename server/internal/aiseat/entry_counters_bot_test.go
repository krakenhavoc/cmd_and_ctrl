package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #4322: a real five-mana board must offer both bodies and cast the
// counter-grown Hydra, with no name or oracle lookup in the policy.
func TestHeuristicCastsKalonianHydraBeforeBurnishedHart(t *testing.T) {
	g := newSettledTable(t, 4322)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards, me.Command.Cards = nil, nil
	for range 5 {
		seedBattlefield(g, me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	}
	hydra, hart := uuid.New(), uuid.New()
	for _, c := range []game.Card{
		{InstanceID: hydra, Name: "Kalonian Hydra", OracleID: "7bd36106-04fe-481f-b16e-e076dcbb183b", TypeLine: "Creature — Hydra", ManaCost: "{3}{G}{G}", Power: 0, Toughness: 0},
		{InstanceID: hart, Name: "Burnished Hart", OracleID: "893fed41-c144-433f-af88-bc7d419b7fb3", TypeLine: "Artifact Creature — Elk", ManaCost: "{3}", Power: 2, Toughness: 2},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		c.AddKnowersAll(seatIDs(g))
		me.Hand.PushTop(c)
	}
	g.BumpLayerVersionForTest()
	moves := legal.EnumerateFor(g, me.ID)
	offered := map[uuid.UUID]bool{}
	for _, m := range moves {
		if m.Kind == legal.KindCast {
			offered[m.Source] = true
		}
	}
	if !offered[hydra] || !offered[hart] {
		t.Fatalf("both casts were not offered: %v", attackLabels(moves))
	}
	in := aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: moves}
	for _, enabled := range []bool{false, true} {
		cfg := heuristic.DefaultConfig()
		cfg.PriceEntryCounters = enabled
		d, err := heuristic.NewWithConfig(cfg).Decide(context.Background(), in)
		if err != nil || d.Index < 0 || d.Index >= len(moves) {
			t.Fatalf("decision %+v: %v", d, err)
		}
		want := hart
		if enabled {
			want = hydra
		}
		mv := moves[d.Index]
		if mv.Kind != legal.KindCast || mv.Source != want {
			t.Fatalf("enabled=%v: chose %s (%s), want %s", enabled, mv.Label, d.Reason, want)
		}
		if enabled {
			if err := actions.Dispatch(g, actions.Action{Type: actions.Type(mv.Type), Player: mv.Player, Caller: me.ID, Params: mv.Params}); err != nil {
				t.Fatalf("dispatch Hydra: %v", err)
			}
		}
	}
	for i := 0; i < 16 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("Hydra did not survive its entry")
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == hydra && c.Counters[game.CounterPlusOne] != 4 {
				t.Errorf("entry counters %v, want four", c.Counters)
			}
		}
	})
}
