package aiseat_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func TestHeuristicPlaysUntappedLandThenTheSpellItUnlocks(t *testing.T) {
	g := newSettledTable(t, 4317)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	mountain, archer := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.Cards = nil
		for _, s := range g.Seats {
			s.Hand.Cards = nil
			s.Command.Cards = nil
		}
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Path of Ancestry", OracleID: "b473e293-59e3-4e04-acf2-622604aeb25f", TypeLine: "Land", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		me.Hand.PushTop(game.Card{InstanceID: mountain, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		me.Hand.PushTop(game.Card{InstanceID: archer, Name: "Firebrand Archer", OracleID: "2e9289d6-dbc6-456d-88cf-d1f534e731d6", TypeLine: "Creature — Human Archer", ManaCost: "{1}{R}", Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	})
	g.BumpLayerVersionForTest()
	made := playHeliodMainPhase(t, g, me, heuristic.New())
	firstLand, cast := uuid.Nil, false
	for _, m := range made {
		if m.move.Kind == legal.KindLand && firstLand == uuid.Nil {
			firstLand = m.move.Source
		}
		if m.move.Kind == legal.KindCast && m.move.Source == archer {
			cast = true
		}
	}
	if firstLand != mountain || !cast || !ogreOnBattlefield(g, archer) {
		t.Fatalf("first land %s (want %s), cast=%v, archer resolved=%v", firstLand, mountain, cast, ogreOnBattlefield(g, archer))
	}
}
