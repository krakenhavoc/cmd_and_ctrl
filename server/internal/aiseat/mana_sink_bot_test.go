package aiseat_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #4559: the real Lion Sash, two Plains and Mother of Runes in hand,
// through the real enumerator, view and policy for one turn. The bot
// casts Mother of Runes first and spends its other {W} on the Sash,
// exiling the opponent's creature card rather than their instant or its
// own creature. Backed out, the Sash's exile outbids the one-drop and
// eats the mana it needed, which is the bug as it was seen on dev.
func TestHeuristicCastsBeforeSinkingManaIntoLionSash(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "sink last", false: "back-out"}[enabled], func(t *testing.T) {
			g := newSettledTable(t, 4559)
			advanceToStep(t, g, game.StepPrecombatMain)
			me := g.Seats[g.Turn.ActiveSeat]
			var opp *game.Player
			for _, s := range g.Seats {
				if s.ID != me.ID {
					opp = s
				}
			}
			everyone := map[uuid.UUID]bool{}
			for _, s := range g.Seats {
				everyone[s.ID] = true
			}
			sash, mother, titan, swords, angel := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
			g.WithWriteLock(func() {
				for _, s := range g.Seats {
					s.Hand.Cards = nil
					s.Command.Cards = nil
				}
				g.Battlefield.PushTop(game.Card{InstanceID: sash, Name: "Lion Sash",
					OracleID: "50330e48-db74-4c5f-a0bc-f9a8607e8f31", TypeLine: "Artifact Creature — Equipment Cat",
					ManaCost: "{1}{W}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, KnownBy: everyone})
				for i := 0; i < 2; i++ {
					g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
						Owner: me.ID, Controller: me.ID, KnownBy: everyone})
				}
				me.Hand.Cards = []game.Card{{InstanceID: mother, Name: "Mother of Runes",
					OracleID: "60433b48-d27f-413c-905c-43839b1943f1", TypeLine: "Creature — Human Cleric", ManaCost: "{W}",
					Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}}}
				me.Graveyard.Cards = []game.Card{{InstanceID: angel, Name: "Baneslayer Angel",
					OracleID: "0e11792b-7fe5-4208-aa0b-e5d09b2b65fe", TypeLine: "Creature — Angel", ManaCost: "{3}{W}{W}",
					Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID, KnownBy: everyone}}
				opp.Graveyard.Cards = []game.Card{
					{InstanceID: titan, Name: "Sun Titan", OracleID: "b2e950fb-cb7e-40a0-a311-5bbdd0477b29",
						TypeLine: "Creature — Giant", ManaCost: "{4}{W}{W}", Power: 6, Toughness: 6,
						Owner: opp.ID, Controller: opp.ID, KnownBy: everyone},
					{InstanceID: swords, Name: "Swords to Plowshares", OracleID: "b1544f21-7e98-461b-aed5-e748b0168c52",
						TypeLine: "Instant", ManaCost: "{W}", Owner: opp.ID, Controller: opp.ID, KnownBy: everyone},
				}
			})
			g.BumpLayerVersionForTest()
			offered := false
			for _, m := range legal.EnumerateFor(g, me.ID) {
				if m.Source == sash && m.Kind == legal.KindActivate {
					offered = true
				}
			}
			if !offered {
				t.Fatal("no real Lion Sash activation offered")
			}
			cfg := heuristic.DefaultConfig()
			cfg.SinkManaLast, cfg.PriceGraveyardExile = enabled, enabled
			made := playHeliodTurn(t, g, me, heuristic.NewWithConfig(cfg))
			if len(made) == 0 {
				t.Fatal("the bot made no move")
			}
			var exiled []uuid.UUID
			for _, m := range made {
				if m.move.Kind == legal.KindActivate && m.move.Source == sash {
					exiled = append(exiled, targetIDs(t, m.move)...)
				}
			}
			cast := g.Battlefield.Contains(mother)
			if !enabled {
				if made[0].move.Source != sash || cast {
					t.Errorf("back-out did not reproduce the sink eating the one-drop's mana: first %q, Mother cast %v", made[0].move.Label, cast)
				}
				return
			}
			if made[0].move.Source != mother || !cast {
				t.Fatalf("first move %q, Mother of Runes on the battlefield %v; want it cast before the sink", made[0].move.Label, cast)
			}
			if len(exiled) != 1 || exiled[0] != titan {
				t.Fatalf("exiled %v; want the one leftover {W} spent on Sun Titan (%s)", exiled, titan)
			}
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == sash && c.Counters["+1/+1"] != 1 {
					t.Errorf("Lion Sash has counters %v; want the +1/+1 counter Sun Titan's exile grows", c.Counters)
				}
			}
		})
	}
}
