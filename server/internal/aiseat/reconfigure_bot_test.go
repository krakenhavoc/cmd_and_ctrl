package aiseat_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #4321: the real Rabbit Battery and both reconfigure halves. With
// nothing worth doing, the bot keeps its mana and reaches its end step.
func TestHeuristicDoesNotWasteManaReconfiguringRabbitBattery(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "priced", false: "back-out"}[enabled], func(t *testing.T) {
			g := newSettledTable(t, 4321)
			advanceToStep(t, g, game.StepPrecombatMain)
			me := g.Seats[g.Turn.ActiveSeat]
			host, battery := uuid.New(), uuid.New()
			g.WithWriteLock(func() {
				for _, s := range g.Seats {
					s.Hand.Cards = nil
					s.Command.Cards = nil
				}
				g.Battlefield.PushTop(game.Card{InstanceID: host, Name: "Test Host", TypeLine: "Creature",
					Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
				g.Battlefield.PushTop(game.Card{InstanceID: battery, Name: "Rabbit Battery",
					OracleID: "c739e180-2f14-41ed-8e7e-50b7df985f35", TypeLine: "Artifact Creature — Equipment Rabbit",
					ManaCost: "{R}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
					AttachedTo: game.TargetRef{Kind: "card", ID: host}, KnownBy: map[uuid.UUID]bool{me.ID: true}})
				for i := 0; i < 5; i++ {
					g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain",
						Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
				}
			})
			g.BumpLayerVersionForTest()
			view := protocol.ViewOfGameFor(g, me.ID.String())
			found := false
			for _, c := range view.Battlefield.Cards {
				if c.InstanceID == battery.String() {
					if len(c.ActivatedAbilities) != 2 {
						t.Fatalf("Rabbit Battery rows: %+v", c.ActivatedAbilities)
					}
					for _, row := range c.ActivatedAbilities {
						if !row.Reconfigure || row.Equip {
							t.Fatalf("reconfigure projection: %+v", row)
						}
					}
					found = true
				}
			}
			if !found {
				t.Fatal("Rabbit Battery missing from view")
			}
			offered := false
			for _, m := range legal.EnumerateFor(g, me.ID) {
				if m.Source == battery && m.Kind == legal.KindActivate {
					offered = true
				}
			}
			if !offered {
				t.Fatal("no real reconfigure activation offered")
			}
			cfg := heuristic.DefaultConfig()
			cfg.PriceReconfigure = enabled
			made := playHeliodTurn(t, g, me, heuristic.NewWithConfig(cfg))
			activations := 0
			for _, m := range made {
				if m.move.Kind == legal.KindActivate && m.move.Source == battery {
					activations++
				}
			}
			if enabled && activations != 0 {
				t.Errorf("made %d pointless reconfigure activations", activations)
			}
			if !enabled && activations == 0 {
				t.Error("back-out did not reproduce the wasted activations")
			}
			if g.Turn.Step != game.StepEnd {
				t.Fatalf("turn did not reach its end: %s", g.Turn.Step)
			}
			if enabled {
				for _, c := range g.Battlefield.Cards {
					if c.Controller == me.ID && c.Name == "Mountain" && c.Tapped {
						t.Error("bot spent mana with nothing to do")
					}
				}
			}
		})
	}
}
