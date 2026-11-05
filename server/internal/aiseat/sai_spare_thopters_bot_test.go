package aiseat_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #4991: the real Sai, Master Thopterist with five tapped Thopter tokens
// and two Islands after combat, through the real enumerator, view and
// policy for the rest of the turn. The bot draws with Sai once,
// sacrificing two Thopters and keeping three. Backed out
// (PriceSpareTokens off), two Thopters cost their whole board value
// against the card and it never activates, which is what the arena showed.
func TestHeuristicCashesInSpareThoptersWithSai(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "spare tokens", false: "back-out"}[enabled], func(t *testing.T) {
			g := newSettledTable(t, 4991)
			advanceToStep(t, g, game.StepPostcombatMain)
			me := g.Seats[g.Turn.ActiveSeat]
			everyone := map[uuid.UUID]bool{}
			for _, s := range g.Seats {
				everyone[s.ID] = true
			}
			sai := uuid.New()
			g.WithWriteLock(func() {
				for _, s := range g.Seats {
					s.Hand.Cards = nil
					s.Command.Cards = nil
				}
				g.Battlefield.PushTop(game.Card{InstanceID: sai, Name: "Sai, Master Thopterist",
					OracleID: "52241b9c-7a69-4176-9234-8bdab09d8e64", TypeLine: "Legendary Creature — Human Artificer",
					ManaCost: "{2}{U}", Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID, KnownBy: everyone})
				for i := 0; i < 5; i++ {
					tk := effects.TokenCard("1/1 colorless Thopter artifact with flying")
					tk.InstanceID, tk.Owner, tk.Controller, tk.KnownBy = uuid.New(), me.ID, me.ID, everyone
					tk.Tapped = true
					g.Battlefield.PushTop(tk)
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
			draws := 0
			for _, m := range made {
				if m.move.Kind == legal.KindActivate && m.move.Source == sai {
					draws++
				}
			}
			thopters := 0
			for _, c := range g.Battlefield.Cards {
				if c.Controller == me.ID && c.Name == "Thopter" {
					thopters++
				}
			}
			want, wantThopters := 1, 3
			if !enabled {
				want, wantThopters = 0, 5
			}
			if draws != want || thopters != wantThopters {
				t.Errorf("Sai activated %d times leaving %d Thopters; want %d and %d", draws, thopters, want, wantThopters)
			}
			if enabled && len(me.Hand.Cards) != 1 {
				t.Errorf("the bot holds %d cards after Sai's draw, want 1", len(me.Hand.Cards))
			}
		})
	}
}
