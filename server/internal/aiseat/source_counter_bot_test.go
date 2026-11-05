package aiseat_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #5083: the real Hangarback Walker with two +1/+1 counters and two
// Islands, in the bot's own second main phase with nothing in hand,
// through the real enumerator, view and policy. Its tap row declares the
// counter it puts on the Walker, so with PriceSourceCounters on the bot
// grows it once and the Walker has three counters. With it off
// (DefaultConfig: the price lost in the arena) the row is the flat
// activation price less the blocker (0.20, under PassThreshold) and the
// bot passes the turn, as it did on its own turn in #4991's arena.
func TestHeuristicGrowsHangarbackWithSpareMana(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "source counters", false: "default"}[enabled], func(t *testing.T) {
			g := newSettledTable(t, 5083)
			advanceToStep(t, g, game.StepPostcombatMain)
			me := g.Seats[g.Turn.ActiveSeat]
			everyone := map[uuid.UUID]bool{}
			for _, s := range g.Seats {
				everyone[s.ID] = true
			}
			walker := uuid.New()
			g.WithWriteLock(func() {
				for _, s := range g.Seats {
					s.Hand.Cards = nil
					s.Command.Cards = nil
				}
				g.Battlefield.PushTop(game.Card{InstanceID: walker, Name: "Hangarback Walker",
					OracleID: "dde55256-5259-44e7-a267-fca45a7f0d04", TypeLine: "Artifact Creature — Construct",
					ManaCost: "{X}{X}", Counters: map[string]int{game.CounterPlusOne: 2},
					Owner: me.ID, Controller: me.ID, KnownBy: everyone})
				for i := 0; i < 2; i++ {
					g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island",
						Owner: me.ID, Controller: me.ID, KnownBy: everyone})
				}
			})
			g.BumpLayerVersionForTest()
			offered := false
			for _, m := range legal.EnumerateFor(g, me.ID) {
				if m.Source == walker && m.Kind == legal.KindActivate {
					offered = true
				}
			}
			if !offered {
				t.Fatal("no real Hangarback Walker activation offered")
			}
			cfg := heuristic.DefaultConfig()
			cfg.PriceSourceCounters = enabled
			made := playHeliodTurn(t, g, me, heuristic.NewWithConfig(cfg))
			activated := 0
			for _, m := range made {
				if m.move.Kind == legal.KindActivate && m.move.Source == walker {
					activated++
				}
			}
			counters := 0
			for i := range g.Battlefield.Cards {
				if c := &g.Battlefield.Cards[i]; c.InstanceID == walker {
					counters = c.Counters[game.CounterPlusOne]
				}
			}
			want, wantCounters := 1, 3
			if !enabled {
				want, wantCounters = 0, 2
			}
			if activated != want || counters != wantCounters {
				t.Errorf("the Walker was activated %d times and has %d counters; want %d and %d", activated, counters, want, wantCounters)
			}
		})
	}
}
