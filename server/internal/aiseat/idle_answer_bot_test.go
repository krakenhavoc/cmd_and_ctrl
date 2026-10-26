package aiseat_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #4318: use the actual catalog declarations and enumerator through a
// postcombat main phase, with no spell on the stack to answer. The
// source is tapped, as after attacking in the reported games.
func TestHeuristicKeepsIdleDefensiveCreatures(t *testing.T) {
	for _, card := range []struct {
		name, oracle     string
		power, toughness int
	}{
		{"Selfless Spirit", "71d785a9-ddc8-472e-a778-a551b444a4bd", 2, 1},
		{"Ranger-Captain of Eos", "cada3481-cc2b-4412-b9b5-0436af53aad2", 3, 3},
	} {
		for _, enabled := range []bool{true, false} {
			name := "fixed"
			if !enabled {
				name = "back-out"
			}
			t.Run(card.name+"/"+name, func(t *testing.T) {
				g := newSettledTable(t, 4318)
				advanceToStep(t, g, game.StepPostcombatMain)
				me := g.Seats[g.Turn.ActiveSeat]
				id := uuid.New()
				g.WithWriteLock(func() {
					for _, s := range g.Seats {
						s.Hand.Cards = nil
						s.Command.Cards = nil
					}
					g.Battlefield.PushTop(game.Card{InstanceID: id, Name: card.name, OracleID: card.oracle,
						TypeLine: "Creature", Power: card.power, Toughness: card.toughness, Tapped: true, Owner: me.ID, Controller: me.ID,
						KnownBy: map[uuid.UUID]bool{me.ID: true}})
					for i := 0; i < 5; i++ {
						g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
							Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
					}
					for i := 0; i < 2; i++ {
						g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Test Ogre", TypeLine: "Creature",
							Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
					}
				})
				g.BumpLayerVersionForTest()
				offered := false
				for _, m := range legal.EnumerateFor(g, me.ID) {
					if m.Source == id && m.Kind == legal.KindActivate {
						offered = true
					}
				}
				if !offered {
					t.Fatal("catalog sacrifice was not offered")
				}
				cfg := heuristic.DefaultConfig()
				cfg.PriceAnswerOnlyRows = enabled
				pol := heuristic.NewWithConfig(cfg)
				made := playHeliodTurn(t, g, me, pol)
				sacrificed := false
				for _, m := range made {
					if m.move.Source == id && m.move.Kind == legal.KindActivate {
						sacrificed = true
					}
				}
				if sacrificed == enabled {
					t.Errorf("sacrificed=%v, defensive pricing enabled=%v", sacrificed, enabled)
				}
				if enabled && !ogreOnBattlefield(g, id) {
					t.Error("the defensive creature left the battlefield")
				}
				if g.Turn.Step != game.StepEnd {
					t.Errorf("stopped at %s, want end step", g.Turn.Step)
				}
			})
		}
	}
}

func TestHeuristicStillSacrificesSelflessSpiritForWrath(t *testing.T) {
	g := newSettledTable(t, 4319)
	advanceToStep(t, g, game.StepPrecombatMain)
	them := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spirit, wrath := uuid.New(), uuid.New()
	ogres := []uuid.UUID{uuid.New(), uuid.New()}
	g.WithWriteLock(func() {
		for _, s := range g.Seats {
			s.Hand.Cards = nil
			s.Command.Cards = nil
		}
		g.Battlefield.PushTop(game.Card{InstanceID: spirit, Name: "Selfless Spirit",
			OracleID: "71d785a9-ddc8-472e-a778-a551b444a4bd", TypeLine: "Creature — Spirit Cleric",
			Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		for _, id := range ogres {
			g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Test Ogre", TypeLine: "Creature",
				Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		}
		for i := 0; i < 4; i++ {
			g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
				Owner: them.ID, Controller: them.ID, KnownBy: map[uuid.UUID]bool{them.ID: true}})
		}
		them.Hand.PushTop(game.Card{InstanceID: wrath, Name: "Wrath of God", OracleID: "34515b16-c9a4-4f98-8c77-416a7a523407",
			TypeLine: "Sorcery", ManaCost: "{2}{W}{W}", Owner: them.ID, Controller: them.ID, KnownBy: map[uuid.UUID]bool{them.ID: true}})
	})
	g.BumpLayerVersionForTest()
	cast := false
	for _, m := range legal.EnumerateFor(g, them.ID) {
		if m.Source == wrath && m.Kind == legal.KindCast {
			dispatchMove(t, g, m)
			cast = true
			break
		}
	}
	if !cast {
		t.Fatal("Wrath was not offered")
	}
	made := playOutTheStack(t, g, me, heuristic.New())
	activated := false
	for _, m := range made {
		if m.Source == spirit && m.Kind == legal.KindActivate {
			activated = true
		}
	}
	if !activated || ogreOnBattlefield(g, spirit) {
		t.Fatal("Spirit was not sacrificed in response")
	}
	for _, id := range ogres {
		if !ogreOnBattlefield(g, id) {
			t.Error("Wrath killed an Ogre despite the Spirit's protection")
		}
	}
}
