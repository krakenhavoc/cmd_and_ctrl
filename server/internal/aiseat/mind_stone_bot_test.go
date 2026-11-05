package aiseat_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #5085: the real Mind Stone with seven Islands and an empty hand, after
// combat, through the real enumerator, view and policy. With
// PriceSpareManaSources on (DefaultConfig) the stone is spare and the bot
// cashes it in for a card. With it off (the arena's
// `heuristic-nosparemana`) the stone costs its board value against the
// card and the bot never activates it, which is what #4991's arena saw.
func TestHeuristicCashesInASpareMindStone(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "default", false: "price off"}[enabled], func(t *testing.T) {
			g := newSettledTable(t, 5085)
			advanceToStep(t, g, game.StepPostcombatMain)
			me := g.Seats[g.Turn.ActiveSeat]
			everyone := map[uuid.UUID]bool{}
			for _, s := range g.Seats {
				everyone[s.ID] = true
			}
			stone := uuid.New()
			g.WithWriteLock(func() {
				for _, s := range g.Seats {
					s.Hand.Cards = nil
					s.Command.Cards = nil
				}
				g.Battlefield.PushTop(game.Card{InstanceID: stone, Name: "Mind Stone",
					OracleID: "c97361b5-af16-4a7b-af85-a429dbaf4ad2", TypeLine: "Artifact",
					ManaCost: "{2}", Owner: me.ID, Controller: me.ID, KnownBy: everyone})
				for i := 0; i < 7; i++ {
					g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island",
						Owner: me.ID, Controller: me.ID, KnownBy: everyone})
				}
			})
			g.BumpLayerVersionForTest()
			offered := false
			for _, m := range legal.EnumerateFor(g, me.ID) {
				if m.Source == stone && m.Kind == legal.KindActivate {
					offered = true
				}
			}
			if !offered {
				t.Fatal("no real Mind Stone activation offered")
			}
			library := len(me.Library.Cards)
			cfg := heuristic.DefaultConfig()
			cfg.PriceSpareManaSources = enabled
			made := playHeliodTurn(t, g, me, heuristic.NewWithConfig(cfg))
			activated := 0
			for _, m := range made {
				if m.move.Kind == legal.KindActivate && m.move.Source == stone {
					activated++
				}
			}
			drawn := library - len(me.Library.Cards)
			if !enabled {
				if activated != 0 || !g.Battlefield.Contains(stone) || drawn != 0 {
					t.Errorf("with the price off the bot activated the stone %d times and drew %d; want it kept, as before #5085", activated, drawn)
				}
				return
			}
			if activated != 1 || g.Battlefield.Contains(stone) {
				t.Fatalf("Mind Stone activated %d times, still on the battlefield %v; want it cashed in once", activated, g.Battlefield.Contains(stone))
			}
			if drawn != 1 {
				t.Errorf("the library lost %d cards; want the one card the stone drew", drawn)
			}
		})
	}
}
