package aiseat_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #4991: the real Sai, Master Thopterist with four Thopter tokens and
// two Islands, after combat, through the real enumerator, view and
// policy. With PriceSpareTokens on, the bot trades two Thopters for a
// card and keeps two. With it off (DefaultConfig: the price lost in the
// arena), the two Thopters cost their bodies and the bot never activates
// Sai, which is what the issue's arena run saw.
func TestHeuristicTradesSpareThoptersToSai(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "spare tokens", false: "default"}[enabled], func(t *testing.T) {
			g := newSettledTable(t, 4991)
			advanceToStep(t, g, game.StepPostcombatMain)
			me := g.Seats[g.Turn.ActiveSeat]
			everyone := map[uuid.UUID]bool{}
			for _, s := range g.Seats {
				everyone[s.ID] = true
			}
			sai := uuid.New()
			var thopters []uuid.UUID
			g.WithWriteLock(func() {
				for _, s := range g.Seats {
					s.Hand.Cards = nil
					s.Command.Cards = nil
				}
				g.Battlefield.PushTop(game.Card{InstanceID: sai, Name: "Sai, Master Thopterist",
					OracleID: "52241b9c-7a69-4176-9234-8bdab09d8e64", TypeLine: "Legendary Creature — Human Artificer",
					ManaCost: "{2}{U}", Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID, KnownBy: everyone})
				for i := 0; i < 4; i++ {
					id := uuid.New()
					thopters = append(thopters, id)
					g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter",
						Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID, KnownBy: everyone})
				}
				for i := 0; i < 2; i++ {
					g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island",
						Owner: me.ID, Controller: me.ID, KnownBy: everyone})
				}
			})
			g.BumpLayerVersionForTest()
			offered := false
			for _, m := range legal.EnumerateFor(g, me.ID) {
				if m.Source == sai && m.Kind == legal.KindActivate {
					offered = true
				}
			}
			if !offered {
				t.Fatal("no real Sai activation offered")
			}
			cfg := heuristic.DefaultConfig()
			cfg.PriceSpareTokens = enabled
			made := playHeliodTurn(t, g, me, heuristic.NewWithConfig(cfg))
			activated := 0
			for _, m := range made {
				if m.move.Kind == legal.KindActivate && m.move.Source == sai {
					activated++
				}
			}
			left := 0
			for _, id := range thopters {
				if g.Battlefield.Contains(id) {
					left++
				}
			}
			if !enabled {
				if activated != 0 || left != 4 {
					t.Errorf("the default activated Sai %d times and kept %d Thopters; want no activation, as before #4991", activated, left)
				}
				return
			}
			if activated != 1 || left != 2 {
				t.Fatalf("Sai activated %d times, %d Thopters left; want one activation keeping two", activated, left)
			}
			if n := len(me.Hand.Cards); n != 1 {
				t.Errorf("hand holds %d cards after the activation; want the card Sai drew", n)
			}
		})
	}
}
